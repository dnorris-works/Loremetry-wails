package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchAnalysisSourcesBook(t *testing.T) {
	root := t.TempDir()
	book, err := ApplyBookTemplate(root, "Tale", "", "Act")
	if err != nil {
		t.Fatal(err)
	}
	got := MatchAnalysisSources(book)
	if got.Kind != "book" {
		t.Fatalf("kind %q", got.Kind)
	}
	ms := got.Role("manuscript")
	if !ms.Present || len(ms.Files) == 0 {
		t.Fatalf("manuscript: %+v", ms)
	}
	if ms.Files[0].Name != "001-Scene-Title.md" {
		t.Fatalf("scene %q", ms.Files[0].Name)
	}
	if !got.Role("characters").Present {
		t.Fatal("characters folder")
	}
	if !got.Role("blurb").Present {
		t.Fatal("blurb file")
	}
	if got.Role("world").Present {
		t.Fatal("world is series-only")
	}
}

func TestMatchAnalysisSourcesSeries(t *testing.T) {
	root := t.TempDir()
	series, err := ApplySeriesTemplate(root, "Cycle")
	if err != nil {
		t.Fatal(err)
	}
	got := MatchAnalysisSources(series)
	if got.Kind != "series" {
		t.Fatalf("kind %q", got.Kind)
	}
	if !got.Role("bible").Present {
		t.Fatal("series bible")
	}
	if !got.Role("characters").Present {
		t.Fatal("series characters")
	}
	bible := got.Role("bible")
	found := false
	for _, f := range bible.Files {
		if f.Name == "Series-Overview.md" {
			found = true
		}
	}
	if !found {
		t.Fatalf("overview in %+v", bible.Files)
	}
}

func TestMatchAnalysisSourcesCaseInsensitive(t *testing.T) {
	root := t.TempDir()
	book := filepath.Join(root, "Tale")
	dir := filepath.Join(book, "01_Manuscript", "01_Chapters")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := MatchAnalysisSources(book)
	if !got.Role("manuscript").Present {
		t.Fatalf("%+v", got.Role("manuscript"))
	}
}
