package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open() (*sql.DB, error) {
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		return nil, nil
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping: %w", err)
	}
	if err := migrate(context.Background(), db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL DEFAULT '',
			plan TEXT NOT NULL DEFAULT 'writer',
			max_concurrent_ai INT NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS device_tokens (
			token_hash TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS credit_ledger (
			user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			balance INT NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS credit_events (
			id BIGSERIAL PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			analysis_id TEXT NOT NULL DEFAULT '',
			delta INT NOT NULL,
			reason TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS ai_jobs (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			analysis_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'queued',
			sources_json TEXT NOT NULL DEFAULT '[]',
			sources_hash TEXT NOT NULL DEFAULT '',
			step INT NOT NULL DEFAULT 0,
			step_total INT NOT NULL DEFAULT 1,
			body TEXT NOT NULL DEFAULT '',
			error_text TEXT NOT NULL DEFAULT '',
			credits_charged INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			started_at TIMESTAMPTZ,
			finished_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS ai_jobs_user_status ON ai_jobs(user_id, status)`,
		`CREATE TABLE IF NOT EXISTS ai_job_steps (
			id BIGSERIAL PRIMARY KEY,
			job_id TEXT NOT NULL REFERENCES ai_jobs(id) ON DELETE CASCADE,
			step_index INT NOT NULL,
			step_name TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending',
			model_tier TEXT NOT NULL DEFAULT '',
			result_text TEXT NOT NULL DEFAULT '',
			prompt_tokens INT NOT NULL DEFAULT 0,
			completion_tokens INT NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS ai_call_log (
			id BIGSERIAL PRIMARY KEY,
			user_id TEXT NOT NULL DEFAULT '',
			job_id TEXT NOT NULL DEFAULT '',
			analysis_id TEXT NOT NULL DEFAULT '',
			step_name TEXT NOT NULL DEFAULT '',
			model TEXT NOT NULL DEFAULT '',
			model_tier TEXT NOT NULL DEFAULT '',
			system_prompt TEXT NOT NULL DEFAULT '',
			user_prompt TEXT NOT NULL DEFAULT '',
			response_body TEXT NOT NULL DEFAULT '',
			prompt_tokens INT NOT NULL DEFAULT 0,
			completion_tokens INT NOT NULL DEFAULT 0,
			total_tokens INT NOT NULL DEFAULT 0,
			duration_ms INT NOT NULL DEFAULT 0,
			http_status INT NOT NULL DEFAULT 0,
			error_text TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS ai_call_log_user ON ai_call_log(user_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS ai_call_log_job ON ai_call_log(job_id)`,
		`CREATE TABLE IF NOT EXISTS analysis_cache (
			cache_key TEXT PRIMARY KEY,
			analysis_id TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	if os.Getenv("LOREMETRY_DEV_SEED") == "1" {
		return seedDev(ctx, db)
	}
	return nil
}

func seedDev(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, plan, max_concurrent_ai) VALUES
			('user_dev', 'dev@loremetry.local', 'writer', 1),
			('user_empty', 'empty@loremetry.local', 'writer', 1)
		ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO credit_ledger (user_id, balance) VALUES
			('user_dev', 100),
			('user_empty', 0)
		ON CONFLICT (user_id) DO NOTHING`)
	if err != nil {
		return err
	}
	for _, row := range []struct {
		hash, uid string
	}{
		{hashToken("dev-token"), "user_dev"},
		{hashToken("empty-token"), "user_empty"},
	} {
		_, err = db.ExecContext(ctx, `
			INSERT INTO device_tokens (token_hash, user_id) VALUES ($1, $2)
			ON CONFLICT (token_hash) DO NOTHING`, row.hash, row.uid)
		if err != nil {
			return err
		}
	}
	return nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}
