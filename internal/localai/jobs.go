package localai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"loremetry/internal/apiserver/analysis"
	"loremetry/internal/apiserver/cache"
	"loremetry/internal/apiserver/models"
	"loremetry/internal/cloud"

	"github.com/google/uuid"
)

// RunStats persists successful analysis durations for next-run estimates.
type RunStats interface {
	AnalysisRunEstimate(analysisID string) (time.Duration, bool)
	RecordAnalysisRun(analysisID string, d time.Duration)
}

// ResultCache stores completed local analysis outputs for instant re-runs.
type ResultCache interface {
	GetAnalysisJobCache(key string) (body string, ok bool)
	PutAnalysisJobCache(key, analysisID, body string)
}

// JobStore tracks in-process local analysis jobs (same shape as cloud.JobStatus).
type JobStore struct {
	mu      sync.Mutex
	jobs    map[string]*cloud.JobStatus
	cancels map[string]context.CancelFunc
	stats   RunStats
	cache   ResultCache
}

func NewJobStore() *JobStore {
	return &JobStore{
		jobs:    map[string]*cloud.JobStatus{},
		cancels: map[string]context.CancelFunc{},
	}
}

func (s *JobStore) SetStats(stats RunStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats = stats
}

func (s *JobStore) SetResultCache(cache ResultCache) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = cache
}

func (s *JobStore) Get(id string) (cloud.JobStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return cloud.JobStatus{}, false
	}
	out := *j
	return out, true
}

// Cancel interrupts a running job. Returns false if the job is unknown or already finished.
func (s *JobStore) Cancel(jobID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	cancel, ok := s.cancels[jobID]
	if !ok {
		j, exists := s.jobs[jobID]
		return exists && (j.Status == "running" || j.Status == "queued")
	}
	if j, ok := s.jobs[jobID]; ok && (j.Status == "running" || j.Status == "queued") {
		j.Status = "cancelled"
		j.Message = "Stopping…"
		j.Error = "Stopped"
	}
	cancel()
	return true
}

func (s *JobStore) put(j *cloud.JobStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *j
	s.jobs[j.JobID] = &cp
}

func (s *JobStore) update(id string, fn func(*cloud.JobStatus)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return
	}
	fn(j)
}

func (s *JobStore) setCancel(id string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancels[id] = cancel
}

func (s *JobStore) clearCancel(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cancels, id)
}

func (s *JobStore) estimate(analysisID string) (time.Duration, bool) {
	s.mu.Lock()
	stats := s.stats
	s.mu.Unlock()
	if stats == nil {
		return 0, false
	}
	return stats.AnalysisRunEstimate(analysisID)
}

func (s *JobStore) record(analysisID string, d time.Duration) {
	s.mu.Lock()
	stats := s.stats
	s.mu.Unlock()
	if stats == nil {
		return
	}
	stats.RecordAnalysisRun(analysisID, d)
}

// StartJob runs analysis locally and returns the initial queued/running status.
func (s *JobStore) StartJob(mgr *Manager, analysisID string, sources []analysis.Source) (cloud.JobStatus, error) {
	if mgr == nil {
		return cloud.JobStatus{}, fmt.Errorf("local AI not available")
	}
	cacheKey := cache.SourcesHash(analysisID, sources)
	s.mu.Lock()
	resultCache := s.cache
	s.mu.Unlock()
	if resultCache != nil {
		if body, ok := resultCache.GetAnalysisJobCache(cacheKey); ok {
			id := "local-" + uuid.NewString()
			st := &cloud.JobStatus{
				JobID:  id,
				Status: "done",
				Body:   body,
				Cached: true,
			}
			s.put(st)
			return *st, nil
		}
	}
	id := "local-" + uuid.NewString()
	estSec := 0
	msg := "Starting local AI…"
	if est, ok := s.estimate(analysisID); ok && est > 0 {
		estSec = int(est.Seconds())
		if estSec < 1 {
			estSec = 1
		}
		msg = fmt.Sprintf("Starting local AI… · typically ~%s", formatElapsed(est))
	}
	st := &cloud.JobStatus{
		JobID:       id,
		Status:      "running",
		Step:        0,
		StepTotal:   1,
		Message:     msg,
		EstimateSec: estSec,
	}
	s.put(st)

	go s.run(mgr, id, analysisID, sources, cacheKey, time.Duration(estSec)*time.Second)
	out := *st
	return out, nil
}

func formatElapsed(d time.Duration) string {
	sec := int(d.Seconds())
	if sec < 0 {
		sec = 0
	}
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	return fmt.Sprintf("%dm %02ds", sec/60, sec%60)
}

func busyMessage(label string, stepElapsed, jobElapsed, estimate time.Duration, multiStep bool) string {
	base := strings.TrimSpace(label)
	base = strings.TrimRight(base, "….")
	if base == "" {
		base = "Working"
	}
	msg := fmt.Sprintf("%s… · %s", base, formatElapsed(stepElapsed))
	if multiStep || estimate > 0 {
		msg += fmt.Sprintf(" · total %s", formatElapsed(jobElapsed))
	}
	if estimate <= 0 {
		return msg
	}
	left := estimate - jobElapsed
	switch {
	case left > 15*time.Second:
		msg += fmt.Sprintf(" · ~%s left", formatElapsed(left))
	case jobElapsed < estimate:
		msg += " · almost done"
	default:
		msg += " · past usual time"
	}
	return msg
}

func isCancelErr(err error, ctx context.Context) bool {
	if ctx != nil && ctx.Err() != nil {
		return true
	}
	return errors.Is(err, context.Canceled)
}

type heartbeatGateway struct {
	inner  models.Gateway
	onTick func(elapsed time.Duration)
}

func (g *heartbeatGateway) Complete(ctx context.Context, tier models.Tier, system, user string) (string, models.Usage, error) {
	done := make(chan struct{})
	start := time.Now()
	if g.onTick != nil {
		g.onTick(0)
		go func() {
			t := time.NewTicker(2 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-done:
					return
				case now := <-t.C:
					g.onTick(now.Sub(start))
				}
			}
		}()
	}
	body, usage, err := g.inner.Complete(ctx, tier, system, user)
	close(done)
	return body, usage, err
}

func (s *JobStore) run(mgr *Manager, jobID, analysisID string, sources []analysis.Source, cacheKey string, estimate time.Duration) {
	jobCtx, jobCancel := context.WithCancel(context.Background())
	s.setCancel(jobID, jobCancel)
	defer func() {
		s.clearCancel(jobID)
		jobCancel()
	}()

	jobStart := time.Now()
	s.update(jobID, func(j *cloud.JobStatus) {
		if estimate > 0 {
			j.EstimateSec = int(estimate.Seconds())
			j.Message = fmt.Sprintf("Starting local AI… · typically ~%s", formatElapsed(estimate))
		} else {
			j.Message = "Starting local AI…"
		}
	})
	startAI := time.Now()
	startDone := make(chan struct{})
	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-startDone:
				return
			case <-jobCtx.Done():
				return
			case <-t.C:
				s.update(jobID, func(j *cloud.JobStatus) {
					if j.Status != "running" {
						return
					}
					j.Message = busyMessage("Starting local AI", time.Since(startAI), time.Since(jobStart), estimate, j.StepTotal > 1)
				})
			}
		}
	}()
	startCtx, startCancel := context.WithTimeout(jobCtx, 120*time.Second)
	err := mgr.EnsureStarted(startCtx)
	startCancel()
	close(startDone)
	if err != nil {
		s.update(jobID, func(j *cloud.JobStatus) {
			if isCancelErr(err, jobCtx) {
				j.Status = "cancelled"
				j.Error = "Stopped"
				j.Message = "Stopped"
				return
			}
			j.Status = "failed"
			j.Error = err.Error()
			j.Message = ""
		})
		return
	}

	var labelMu sync.Mutex
	currentLabel := "Preparing analysis…"
	stepStart := time.Now()

	setLabel := func(label string) {
		labelMu.Lock()
		currentLabel = label
		stepStart = time.Now()
		labelMu.Unlock()
	}

	gw := &heartbeatGateway{
		inner: mgr.Gateway(),
		onTick: func(stepElapsed time.Duration) {
			labelMu.Lock()
			label := currentLabel
			started := stepStart
			labelMu.Unlock()
			if stepElapsed <= 0 {
				stepElapsed = time.Since(started)
			}
			jobElapsed := time.Since(jobStart)
			s.update(jobID, func(j *cloud.JobStatus) {
				if j.Status != "running" {
					return
				}
				j.Message = busyMessage(label, stepElapsed, jobElapsed, estimate, j.StepTotal > 1)
			})
		},
	}

	runner := analysis.Runner{
		Gateway: gw,
		OnProgress: func(step, total int, label string) {
			if total < 1 {
				total = 1
			}
			if step < 1 {
				step = 1
			}
			if step > total {
				step = total
			}
			setLabel(label)
			jobElapsed := time.Since(jobStart)
			s.update(jobID, func(j *cloud.JobStatus) {
				if j.Status == "cancelled" {
					return
				}
				j.Status = "running"
				j.Step = step
				j.StepTotal = total
				j.Message = busyMessage(label, 0, jobElapsed, estimate, total > 1)
			})
		},
	}
	setLabel("Preparing analysis…")
	s.update(jobID, func(j *cloud.JobStatus) {
		if j.Status == "cancelled" {
			return
		}
		j.Message = busyMessage("Preparing analysis", 0, time.Since(jobStart), estimate, j.StepTotal > 1)
	})
	if jobCtx.Err() != nil {
		s.update(jobID, func(j *cloud.JobStatus) {
			j.Status = "cancelled"
			j.Error = "Stopped"
			j.Message = "Stopped"
		})
		return
	}
	body, err := runner.Run(jobCtx, analysisID, sources)
	elapsed := time.Since(jobStart)
	s.update(jobID, func(j *cloud.JobStatus) {
		if err != nil {
			if isCancelErr(err, jobCtx) || j.Status == "cancelled" {
				j.Status = "cancelled"
				j.Error = "Stopped"
				j.Message = "Stopped"
				return
			}
			j.Status = "failed"
			j.Error = err.Error()
			j.Message = ""
			return
		}
		j.Status = "done"
		j.Body = body
		if j.StepTotal < 1 {
			j.StepTotal = 1
		}
		j.Step = j.StepTotal
		j.Message = fmt.Sprintf("Done in %s", formatElapsed(elapsed))
	})
	if err == nil {
		s.record(analysisID, elapsed)
		s.mu.Lock()
		resultCache := s.cache
		s.mu.Unlock()
		if resultCache != nil && cacheKey != "" && body != "" {
			resultCache.PutAnalysisJobCache(cacheKey, analysisID, body)
		}
	}
}
