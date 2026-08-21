package analysis

import (
	"path/filepath"
	"strings"
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

func TestTruncateExcerptDoesNotAppendRemainder(t *testing.T) {
	var b strings.Builder
	b.WriteString(strings.Repeat("word ", 4000))
	b.WriteString("\n\nChapter 2\n")
	b.WriteString(strings.Repeat("more ", 20000))
	got := TruncateExcerpt(b.String(), 5000)
	if len(got) > 5200 {
		t.Fatalf("truncated too large: %d", len(got))
	}
	if strings.Contains(got, "Chapter 2") {
		t.Fatal("truncated must not append later chapters")
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

	if ProfileFor("analysis") != ProfileByChapter {
		t.Fatal("expected by_chapter profile for analysis")
	}
	if ProfileFor("genre_analysis") != ProfileByChapter {
		t.Fatal("expected by_chapter profile for genre_analysis")
	}
	if ProfileFor("ai_beta_reader") != ProfileByChapter {
		t.Fatal("expected by_chapter profile for ai_beta_reader")
	}
	if ProfileFor("hook_strength") != ProfileSingle {
		t.Fatal("expected single profile for hook_strength")
	}
}
