package queue

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"loremetry/internal/apiserver/analysis"
	"loremetry/internal/apiserver/cache"
	"loremetry/internal/apiserver/ledger"
	"loremetry/internal/apiserver/models"
)

type Status string

const (
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type Job struct {
	ID             string           `json:"job_id"`
	UserID         string           `json:"-"`
	AnalysisID     string           `json:"analysis_id"`
	Status         Status           `json:"status"`
	Step           int              `json:"step"`
	StepTotal      int              `json:"step_total"`
	Body           string           `json:"body,omitempty"`
	Error          string           `json:"error,omitempty"`
	Credits        int              `json:"credits"`
	Cached         bool             `json:"cached,omitempty"`
	Sources        []analysis.Source `json:"-"`
	SourcesHash    string           `json:"-"`
	CreditsCharged int              `json:"-"`
}

type Service struct {
	mu       sync.Mutex
	jobs     map[string]*Job
	db       *sql.DB
	ledger   *ledger.Ledger
	cache    *cache.Store
	gateway  models.Gateway
	runner   analysis.Runner
	workers  *Workers
}

func NewMemory(ledger *ledger.Ledger, cacheStore *cache.Store, gw models.Gateway) *Service {
	s := &Service{
		jobs:    map[string]*Job{},
		ledger:  ledger,
		cache:   cacheStore,
		gateway: gw,
		runner:  analysis.Runner{Gateway: gw},
	}
	s.workers = NewWorkers(s)
	s.workers.Start()
	return s
}

func NewPostgres(db *sql.DB, ledger *ledger.Ledger, cacheStore *cache.Store, gw models.Gateway) *Service {
	s := &Service{
		db:      db,
		ledger:  ledger,
		cache:   cacheStore,
		gateway: gw,
		runner:  analysis.Runner{Gateway: gw},
	}
	s.workers = NewWorkers(s)
	s.workers.Start()
	return s
}

func (s *Service) ActiveCount(userID string) int {
	if s.db != nil {
		var n int
		_ = s.db.QueryRow(`
			SELECT count(*) FROM ai_jobs
			WHERE user_id = $1 AND status IN ('queued','running')`, userID).Scan(&n)
		return n
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, j := range s.jobs {
		if j.UserID == userID && (j.Status == StatusQueued || j.Status == StatusRunning) {
			n++
		}
	}
	return n
}

func (s *Service) Enqueue(userID, analysisID string, sources []analysis.Source, maxConcurrent int) (*Job, error) {
	if maxConcurrent <= 0 {
		return nil, ErrPlanRequired
	}
	if s.ActiveCount(userID) >= maxConcurrent {
		return nil, ErrConcurrentLimit
	}
	hash := cache.SourcesHash(analysisID, sources)
	if e, ok := s.cache.Get(hash); ok {
		return &Job{
			ID:         newID(),
			UserID:     userID,
			AnalysisID: analysisID,
			Status:     StatusDone,
			Step:       StepTotal(analysisID),
			StepTotal:  StepTotal(analysisID),
			Body:       e.Body,
			Credits:    s.ledger.Balance(userID),
			Cached:     true,
		}, nil
	}
	left, ok := s.ledger.Debit(userID, analysisID)
	if !ok {
		return nil, ErrCreditsEmpty
	}
	j := &Job{
		ID:             newID(),
		UserID:         userID,
		AnalysisID:     analysisID,
		Status:         StatusQueued,
		Step:           0,
		StepTotal:      StepTotal(analysisID),
		Sources:        sources,
		SourcesHash:    hash,
		Credits:        left,
		CreditsCharged: s.ledger.CreditCost(analysisID),
	}
	if s.db != nil {
		if err := s.insertDB(j); err != nil {
			s.ledger.Refund(userID, analysisID)
			return nil, err
		}
	} else {
		s.mu.Lock()
		s.jobs[j.ID] = j
		s.mu.Unlock()
	}
	return j, nil
}

func StepTotal(analysisID string) int {
	return analysis.StepTotal(analysis.ProfileFor(analysisID))
}

func (s *Service) Get(userID, jobID string) (*Job, error) {
	if s.db != nil {
		return s.getDB(userID, jobID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[jobID]
	if !ok || j.UserID != userID {
		return nil, ErrNotFound
	}
	out := *j
	return &out, nil
}

func (s *Service) Wait(ctx context.Context, userID, jobID string) (*Job, error) {
	for {
		j, err := s.Get(userID, jobID)
		if err != nil {
			return nil, err
		}
		if j.Status == StatusDone || j.Status == StatusFailed {
			return j, nil
		}
		select {
		case <-ctx.Done():
			return j, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (s *Service) claimNext() (*Job, error) {
	if s.db != nil {
		return s.claimDB()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, j := range s.jobs {
		if j.Status == StatusQueued {
			j.Status = StatusRunning
			j.Step = 1
			out := *j
			return &out, nil
		}
	}
	return nil, nil
}

func (s *Service) finish(j *Job, body string, err error) {
	if err != nil {
		j.Status = StatusFailed
		j.Error = err.Error()
		s.ledger.Refund(j.UserID, j.AnalysisID)
	} else {
		j.Status = StatusDone
		j.Body = body
		j.Step = j.StepTotal
		s.cache.Put(j.SourcesHash, j.AnalysisID, body)
	}
	if s.db != nil {
		s.finishDB(j)
		return
	}
	s.mu.Lock()
	if cur, ok := s.jobs[j.ID]; ok {
		cur.Status = j.Status
		cur.Body = j.Body
		cur.Error = j.Error
		cur.Step = j.Step
	}
	s.mu.Unlock()
}

func (s *Service) runJob(j *Job) {
	ctx := context.Background()
	j.Step = 1
	body, err := s.runner.Run(ctx, j.AnalysisID, j.Sources)
	for step := 2; step <= j.StepTotal; step++ {
		j.Step = step
	}
	s.finish(j, body, err)
}

func (s *Service) insertDB(j *Job) error {
	raw, _ := json.Marshal(j.Sources)
	_, err := s.db.Exec(`
		INSERT INTO ai_jobs (id, user_id, analysis_id, status, sources_json, sources_hash, step, step_total, credits_charged)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		j.ID, j.UserID, j.AnalysisID, j.Status, string(raw), j.SourcesHash, j.Step, j.StepTotal, j.CreditsCharged)
	return err
}

func (s *Service) getDB(userID, jobID string) (*Job, error) {
	var j Job
	var raw string
	err := s.db.QueryRow(`
		SELECT id, user_id, analysis_id, status, sources_json, step, step_total, body, error_text, credits_charged
		FROM ai_jobs WHERE id = $1 AND user_id = $2`, jobID, userID).
		Scan(&j.ID, &j.UserID, &j.AnalysisID, &j.Status, &raw, &j.Step, &j.StepTotal, &j.Body, &j.Error, &j.CreditsCharged)
	if err != nil {
		return nil, ErrNotFound
	}
	_ = json.Unmarshal([]byte(raw), &j.Sources)
	j.Credits = s.ledger.Balance(userID)
	return &j, nil
}

func (s *Service) claimDB() (*Job, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var j Job
	var raw string
	err = tx.QueryRow(`
		SELECT id, user_id, analysis_id, sources_json, sources_hash, step_total, credits_charged
		FROM ai_jobs WHERE status = 'queued'
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1`).Scan(&j.ID, &j.UserID, &j.AnalysisID, &raw, &j.SourcesHash, &j.StepTotal, &j.CreditsCharged)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(raw), &j.Sources)
	j.Status = StatusRunning
	j.Step = 1
	_, err = tx.Exec(`UPDATE ai_jobs SET status = 'running', step = 1, started_at = now() WHERE id = $1`, j.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &j, nil
}

func (s *Service) finishDB(j *Job) {
	status := string(j.Status)
	_, _ = s.db.Exec(`
		UPDATE ai_jobs SET status = $1, step = $2, body = $3, error_text = $4, finished_at = now()
		WHERE id = $5`, status, j.Step, j.Body, j.Error, j.ID)
}

func newID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var (
	ErrNotFound        = fmt.Errorf("job not found")
	ErrCreditsEmpty    = fmt.Errorf("credits empty")
	ErrConcurrentLimit = fmt.Errorf("concurrent limit")
	ErrPlanRequired    = fmt.Errorf("plan required")
)

func IsConcurrentLimit(err error) bool {
	return err != nil && strings.Contains(err.Error(), "concurrent limit")
}

func IsCreditsEmpty(err error) bool {
	return err != nil && strings.Contains(err.Error(), "credits empty")
}
