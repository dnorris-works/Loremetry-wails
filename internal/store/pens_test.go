package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsurePenDirReusesExistingName(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "Adrian Reeve")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := EnsurePenDir(root, "Adrian Reeve")
	if err != nil {
		t.Fatal(err)
	}
	if got != existing {
		t.Fatalf("got %s want %s", got, existing)
	}
	hyphen, err := EnsurePenDir(root, "Adrian-Reeve")
	if err != nil {
		t.Fatal(err)
	}
	if hyphen != existing {
		t.Fatalf("hyphen alias created %s", hyphen)
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 1 {
		t.Fatalf("created extra pen folders: %v", entries)
	}
}
