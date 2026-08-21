package store

import (
	"database/sql"
	"time"
)

// AnalysisRunEstimate returns the rolling average duration for an analysis, if known.
func (s *Store) AnalysisRunEstimate(analysisID string) (time.Duration, bool) {
	if s == nil || s.DB == nil || analysisID == "" {
		return 0, false
	}
	var avgMs int64
	err := s.DB.QueryRow(
		`SELECT avg_ms FROM analysis_run_stats WHERE analysis_id = ? AND run_count > 0 AND avg_ms > 0`,
		analysisID,
	).Scan(&avgMs)
	if err == sql.ErrNoRows || avgMs <= 0 {
		return 0, false
	}
	if err != nil {
		return 0, false
	}
	return time.Duration(avgMs) * time.Millisecond, true
}

// RecordAnalysisRun stores a successful local AI run duration for later estimates.
func (s *Store) RecordAnalysisRun(analysisID string, d time.Duration) {
	if s == nil || s.DB == nil || analysisID == "" {
		return
	}
	ms := d.Milliseconds()
	if ms < 1 {
		ms = 1
	}
	_, _ = s.DB.Exec(`
		INSERT INTO analysis_run_stats (analysis_id, run_count, last_ms, avg_ms, updated_at)
		VALUES (?, 1, ?, ?, datetime('now'))
		ON CONFLICT(analysis_id) DO UPDATE SET
			run_count = analysis_run_stats.run_count + 1,
			last_ms = excluded.last_ms,
			avg_ms = (analysis_run_stats.avg_ms * analysis_run_stats.run_count + excluded.last_ms)
				/ (analysis_run_stats.run_count + 1),
			updated_at = datetime('now')
	`, analysisID, ms, ms)
}
