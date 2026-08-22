package store

import (
	"path/filepath"
	"testing"

	appdb "loremetry/internal/db"
)

func TestAnalysisJobCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	s := &Store{DB: conn, UserID: 1}
	key := "abc123"
	if body, ok := s.GetAnalysisJobCache(key); ok {
		t.Fatalf("unexpected hit: %q", body)
	}
	s.PutAnalysisJobCache(key, "pov_discipline", "# Report\n")
	body, ok := s.GetAnalysisJobCache(key)
	if !ok || body != "# Report\n" {
		t.Fatalf("got %q ok=%v", body, ok)
	}
	s.PutAnalysisJobCache(key, "pov_discipline", "# Updated\n")
	body, ok = s.GetAnalysisJobCache(key)
	if !ok || body != "# Updated\n" {
		t.Fatalf("update got %q ok=%v", body, ok)
	}
}
