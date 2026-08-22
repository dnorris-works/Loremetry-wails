package store

import "database/sql"

// GetAnalysisJobCache returns a cached local analysis result keyed by SourcesHash.
func (s *Store) GetAnalysisJobCache(key string) (body string, ok bool) {
	if s == nil || s.DB == nil || key == "" {
		return "", false
	}
	err := s.DB.QueryRow(`SELECT body FROM analysis_job_cache WHERE cache_key = ?`, key).Scan(&body)
	if err != nil {
		return "", false
	}
	return body, body != ""
}

// PutAnalysisJobCache stores a completed local analysis result.
func (s *Store) PutAnalysisJobCache(key, analysisID, body string) {
	if s == nil || s.DB == nil || key == "" || body == "" {
		return
	}
	_, _ = s.DB.Exec(`
		INSERT INTO analysis_job_cache (cache_key, analysis_id, body) VALUES (?, ?, ?)
		ON CONFLICT(cache_key) DO UPDATE SET
			analysis_id = excluded.analysis_id,
			body = excluded.body,
			created_at = datetime('now')`,
		key, analysisID, body)
}

// EnsureAnalysisJobCacheTable creates the cache table on older DBs opened without full migrate.
func EnsureAnalysisJobCacheTable(db *sql.DB) error {
	if db == nil {
		return nil
	}
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS analysis_job_cache (
			cache_key    TEXT PRIMARY KEY,
			analysis_id  TEXT NOT NULL,
			body         TEXT NOT NULL,
			created_at   TEXT NOT NULL DEFAULT (datetime('now'))
		)`)
	return err
}
