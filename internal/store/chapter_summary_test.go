package store

import (
	"os"
	"path/filepath"
	"testing"

	appdb "loremetry/internal/db"
)

func TestFirstTwoSentences(t *testing.T) {
	got := firstTwoSentences("Hello world. Second sentence! Third.")
	want := "Hello world. Second sentence!"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestListChapterSummaryChaptersSeedsDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	s := &Store{DB: conn, UserID: 1}

	root := t.TempDir()
	chapDir := filepath.Join(root, "01_Manuscript", "01_Chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chapDir, "01-Open.md"), []byte("Alpha one. Beta two. Gamma three."), 0o644); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListChapterSummaryChapters(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len=%d", len(list))
	}
	if list[0].Description != "Alpha one. Beta two." {
		t.Fatalf("desc=%q", list[0].Description)
	}

	// Second call should reuse DB value even if file changes.
	if err := os.WriteFile(filepath.Join(chapDir, "01-Open.md"), []byte("Changed forever."), 0o644); err != nil {
		t.Fatal(err)
	}
	list2, err := s.ListChapterSummaryChapters(root)
	if err != nil {
		t.Fatal(err)
	}
	if list2[0].Description != "Alpha one. Beta two." {
		t.Fatalf("expected persisted desc, got %q", list2[0].Description)
	}
}

func TestFilterManuscriptByRels(t *testing.T) {
	blobs := []RoleText{
		{Role: "manuscript", Rel: "01-A.md", Name: "01-A.md", Text: "a"},
		{Role: "manuscript", Rel: "02-B.md", Name: "02-B.md", Text: "b"},
		{Role: "plot", Rel: "plot.md", Name: "plot.md", Text: "p"},
	}
	got := FilterManuscriptByRels(blobs, []string{"02-B.md"})
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Rel != "02-B.md" || got[1].Role != "plot" {
		t.Fatalf("%+v", got)
	}
}
