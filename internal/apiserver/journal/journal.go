package journal

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"loremetry/internal/apiserver/models"
)

type Entry struct {
	UserID           string
	JobID            string
	AnalysisID       string
	StepName         string
	Model            string
	ModelTier        string
	SystemPrompt     string
	UserPrompt       string
	ResponseBody     string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	DurationMs       int
	HTTPStatus       int
	ErrorText        string
}

type Logger struct {
	db *sql.DB
	mu sync.Mutex
	// In-memory fallback when no Postgres
	entries []Entry
}

func New(db *sql.DB) *Logger {
	return &Logger{db: db}
}

func (l *Logger) Log(e Entry) {
	if l.db != nil {
		_, err := l.db.Exec(`
			INSERT INTO ai_call_log
				(user_id, job_id, analysis_id, step_name, model, model_tier,
				 system_prompt, user_prompt, response_body,
				 prompt_tokens, completion_tokens, total_tokens,
				 duration_ms, http_status, error_text)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
			e.UserID, e.JobID, e.AnalysisID, e.StepName, e.Model, e.ModelTier,
			e.SystemPrompt, e.UserPrompt, e.ResponseBody,
			e.PromptTokens, e.CompletionTokens, e.TotalTokens,
			e.DurationMs, e.HTTPStatus, e.ErrorText)
		if err != nil {
			log.Printf("journal: %v", err)
		}
		return
	}
	l.mu.Lock()
	l.entries = append(l.entries, e)
	l.mu.Unlock()
}

// Entries returns in-memory entries (testing/no-Postgres only).
func (l *Logger) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// LoggingGateway wraps a Gateway and logs every call to the journal.
type LoggingGateway struct {
	Inner      models.Gateway
	Journal    *Logger
	UserID     string
	JobID      string
	AnalysisID string
	StepName   string
}

func (g *LoggingGateway) Complete(ctx context.Context, tier models.Tier, system, user string) (string, models.Usage, error) {
	tierName := "fast"
	if tier == models.TierStrong {
		tierName = "strong"
	}
	modelName := ""
	if oc, ok := g.Inner.(*models.OpenAICompat); ok {
		if tier == models.TierFast {
			modelName = oc.FastModel
		} else {
			modelName = oc.StrongModel
		}
	}

	start := time.Now()
	body, usage, err := g.Inner.Complete(ctx, tier, system, user)
	dur := time.Since(start)

	e := Entry{
		UserID:           g.UserID,
		JobID:            g.JobID,
		AnalysisID:       g.AnalysisID,
		StepName:         g.StepName,
		Model:            modelName,
		ModelTier:        tierName,
		SystemPrompt:     system,
		UserPrompt:       user,
		ResponseBody:     body,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		DurationMs:       int(dur.Milliseconds()),
		HTTPStatus:       200,
	}
	if err != nil {
		e.HTTPStatus = 0
		e.ErrorText = err.Error()
	}
	g.Journal.Log(e)
	return body, usage, err
}

func (g *LoggingGateway) WithStep(step string) *LoggingGateway {
	return &LoggingGateway{
		Inner:      g.Inner,
		Journal:    g.Journal,
		UserID:     g.UserID,
		JobID:      g.JobID,
		AnalysisID: g.AnalysisID,
		StepName:   step,
	}
}
