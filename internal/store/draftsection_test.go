package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenameDraftSectionsPartToAct(t *testing.T) {
	root := t.TempDir()
	from := filepath.Join(root, "01_Manuscript", "01_Chapters", "Part-01")
	if err := os.MkdirAll(from, 0o755); err != nil {
		t.Fatal(err)
	}
	n, err := RenameDraftSections(root, "Act")
	if err != nil || n != 1 {
		t.Fatalf("renamed %d %v", n, err)
	}
	if _, err := os.Stat(filepath.Join(root, "01_Manuscript", "01_Chapters", "Act-01")); err != nil {
		t.Fatal(err)
	}
}

func TestApplyBookTemplateUsesAct(t *testing.T) {
	root := t.TempDir()
	book, err := ApplyBookTemplate(root, "Tale", "", "Act")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(book, "01_Manuscript", "01_Chapters", "Act-01")); err != nil {
		t.Fatal(err)
	}
}
