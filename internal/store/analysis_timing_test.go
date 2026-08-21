package store

import (
	"path/filepath"
	"testing"
	"time"

	appdb "loremetry/internal/db"
)

func TestAnalysisRunEstimateRollingAvg(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	s := &Store{DB: conn, UserID: 1}

	if _, ok := s.AnalysisRunEstimate("pov_discipline"); ok {
		t.Fatal("expected no estimate yet")
	}
	s.RecordAnalysisRun("pov_discipline", 10*time.Minute)
	est, ok := s.AnalysisRunEstimate("pov_discipline")
	if !ok {
		t.Fatal("expected estimate")
	}
	if est < 9*time.Minute || est > 11*time.Minute {
		t.Fatalf("got %v", est)
	}
	s.RecordAnalysisRun("pov_discipline", 2*time.Minute)
	est, ok = s.AnalysisRunEstimate("pov_discipline")
	if !ok {
		t.Fatal("expected estimate")
	}
	// (10m + 2m) / 2 = 6m
	if est < 5*time.Minute || est > 7*time.Minute {
		t.Fatalf("rolling avg got %v", est)
	}
}
