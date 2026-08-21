package localai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"loremetry/internal/apiserver/analysis"
	"loremetry/internal/apiserver/models"
	"loremetry/internal/cloud"

	"github.com/google/uuid"
)

// JobStore tracks in-process local analysis jobs (same shape as cloud.JobStatus).
type JobStore struct {
	mu   sync.Mutex
	jobs map[string]*cloud.JobStatus
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: map[string]*cloud.JobStatus{}}
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

// StartJob runs analysis locally and returns the initial queued/running status.
func (s *JobStore) StartJob(mgr *Manager, analysisID string, sources []analysis.Source) (cloud.JobStatus, error) {
	if mgr == nil {
		return cloud.JobStatus{}, fmt.Errorf("local AI not available")
	}
	id := "local-" + uuid.NewString()
	st := &cloud.JobStatus{
		JobID:     id,
		Status:    "running",
		Step:      0,
		StepTotal: 1,
		Message:   "Starting local AI…",
	}
	s.put(st)

	go s.run(mgr, id, analysisID, sources)
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

func busyMessage(label string, elapsed time.Duration) string {
	base := strings.TrimSpace(label)
	base = strings.TrimRight(base, "….")
	if base == "" {
		base = "Working"
	}
	return fmt.Sprintf("%s… · %s", base, formatElapsed(elapsed))
}

type heartbeatGateway struct {
	inner models.Gateway
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

func (s *JobStore) run(mgr *Manager, jobID, analysisID string, sources []analysis.Source) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	s.update(jobID, func(j *cloud.JobStatus) {
		j.Message = "Starting local AI…"
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
			case <-t.C:
				s.update(jobID, func(j *cloud.JobStatus) {
					if j.Status != "running" {
						return
					}
					j.Message = busyMessage("Starting local AI", time.Since(startAI))
				})
			}
		}
	}()
	err := mgr.EnsureStarted(ctx)
	close(startDone)
	if err != nil {
		s.update(jobID, func(j *cloud.JobStatus) {
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
		onTick: func(elapsed time.Duration) {
			labelMu.Lock()
			label := currentLabel
			started := stepStart
			labelMu.Unlock()
			if elapsed <= 0 {
				elapsed = time.Since(started)
			}
			s.update(jobID, func(j *cloud.JobStatus) {
				if j.Status != "running" {
					return
				}
				j.Message = busyMessage(label, elapsed)
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
			s.update(jobID, func(j *cloud.JobStatus) {
				j.Status = "running"
				j.Step = step
				j.StepTotal = total
				j.Message = busyMessage(label, 0)
			})
		},
	}
	setLabel("Preparing analysis…")
	s.update(jobID, func(j *cloud.JobStatus) {
		j.Message = busyMessage("Preparing analysis", 0)
	})
	body, err := runner.Run(context.Background(), analysisID, sources)
	s.update(jobID, func(j *cloud.JobStatus) {
		if err != nil {
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
		j.Message = "Done"
	})
}
