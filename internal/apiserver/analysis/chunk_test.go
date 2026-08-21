package analysis

import (
	"path/filepath"
	"testing"

	appdb "loremetry/internal/db"
	"loremetry/internal/store"
)

func TestSplitManuscriptBySize(t *testing.T) {
	text := string(make([]byte, 25000))
	for i := range text {
		text = text[:i] + "a" + text[i+1:]
	}
	chunks := SplitManuscript(text, 10000)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
}

func TestProfileForChunked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	store.SetCatalogDB(conn)

	if ProfileFor("analysis") != ProfileChunked {
		t.Fatal("expected chunked profile for analysis")
	}
	if ProfileFor("genre_analysis") != ProfileSingle {
		t.Fatal("expected single profile for genre_analysis")
	}
}
