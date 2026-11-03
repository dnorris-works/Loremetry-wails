package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncDiskHashesDetectsMove(t *testing.T) {
	s := testStore(t)
	root := t.TempDir()
	old := filepath.Join(root, "01_Manuscript", "01_Chapters")
	next := filepath.Join(root, "01_Manuscript", "Chapters")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "01-Open.md"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := s.SyncDiskHashes(root)
	if err != nil || len(changed) == 0 {
		t.Fatalf("first scan: %v %v", changed, err)
	}
	changed, err = s.SyncDiskHashes(root)
	if err != nil || len(changed) != 0 {
		t.Fatalf("stable: %v %v", changed, err)
	}
	if err := os.MkdirAll(next, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(old, "01-Open.md"), filepath.Join(next, "01-Open.md")); err != nil {
		t.Fatal(err)
	}
	changed, err = s.SyncDiskHashes(root)
	if err != nil || len(changed) == 0 {
		t.Fatalf("move: %v %v", changed, err)
	}
}
