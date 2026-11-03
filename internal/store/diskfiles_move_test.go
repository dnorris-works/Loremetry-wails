package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveDiskFileBetweenFolders(t *testing.T) {
	root := t.TempDir()
	main := filepath.Join(root, "Main")
	supp := filepath.Join(root, "Supporting")
	if err := os.MkdirAll(main, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(supp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(main, "Zoe.md"), []byte("roommate"), 0o644); err != nil {
		t.Fatal(err)
	}
	moved, err := MoveDiskFile(main, "Zoe.md", supp)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Dir != supp || moved.Name != "Zoe.md" {
		t.Fatalf("%+v", moved)
	}
	if _, err := os.Stat(filepath.Join(main, "Zoe.md")); !os.IsNotExist(err) {
		t.Fatal("source should be gone")
	}
	if _, err := os.Stat(filepath.Join(supp, "Zoe.md")); err != nil {
		t.Fatal(err)
	}
}

func TestCharacterImportanceFromRel(t *testing.T) {
	if got := characterImportanceFromRel("Main/Zoe.md"); got != "Main" {
		t.Fatalf("got %q", got)
	}
	if got := characterImportanceFromRel("Supporting/Rachel.md"); got != "Supporting" {
		t.Fatalf("got %q", got)
	}
	if got := characterImportanceFromRel("Minor/Extra.md"); got != "Minor" {
		t.Fatalf("got %q", got)
	}
	if got := characterImportanceFromRel("Zoe.md"); got != "" {
		t.Fatalf("root file should be empty, got %q", got)
	}
}
