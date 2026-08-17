package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplySeriesAndBookTemplates(t *testing.T) {
	parent := t.TempDir()
	seriesRoot, err := ApplySeriesTemplate(parent, "The Cycle")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(seriesRoot) != "The Cycle" {
		t.Fatalf("series folder %s", seriesRoot)
	}
	for _, rel := range []string{
		"00_Series-Bible/Series-Overview.md",
		"00_Series-Bible/Characters/Main",
		"Books",
		"98_Series-Admin/Publication-Schedule.md",
	} {
		if _, err := os.Stat(filepath.Join(seriesRoot, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	bookRoot, err := ApplyBookTemplate(filepath.Join(seriesRoot, "Books"), "Book One", "The Cycle", "Part")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(bookRoot) != "01_Book One" {
		t.Fatalf("book folder %q", filepath.Base(bookRoot))
	}
	if _, err := os.Stat(filepath.Join(bookRoot, "01_Manuscript", "00_Current-Draft", "Part-01", "Ch-001", "001-Scene-Title.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(bookRoot, "02_Story-Elements", "Characters", "New-Characters-Introduced.md")); err != nil {
		t.Fatal(err)
	}
	alone, err := ApplyBookTemplate(parent, "Standalone", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(alone) != "Standalone" {
		t.Fatalf("standalone %s", alone)
	}
	if _, err := os.Stat(filepath.Join(alone, "02_Story-Elements", "Characters", "New-Characters-Introduced.md")); !os.IsNotExist(err) {
		t.Fatal("series_only file should be skipped for a standalone book")
	}
}
