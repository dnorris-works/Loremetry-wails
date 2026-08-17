package store

import "testing"

func TestUISessionRoundTrip(t *testing.T) {
	s := testStore(t)
	got, err := s.GetUISession()
	if err != nil {
		t.Fatal(err)
	}
	if !got.RestoreOpen || got.Selection.Type != "empty" {
		t.Fatalf("default: %+v", got)
	}
	got, err = s.SetUISelection("/tmp/dir", "Note.md")
	if err != nil {
		t.Fatal(err)
	}
	if got.Selection.Type != "file" || got.Selection.Name != "Note.md" {
		t.Fatalf("select: %+v", got)
	}
	got, err = s.ToggleOpenSeries("/series")
	if err != nil || len(got.OpenSeries) != 1 {
		t.Fatalf("toggle: %+v %v", got, err)
	}
	got, err = s.ToggleOpenFolder("/book/00_Admin")
	if err != nil || len(got.OpenFolders) != 1 {
		t.Fatalf("folder: %+v %v", got, err)
	}
	got, err = s.SetLastPen("Ada")
	if err != nil || got.LastPen != "Ada" {
		t.Fatalf("pen: %+v %v", got, err)
	}
	got, err = s.SetRestoreOpen(false)
	if err != nil || got.RestoreOpen {
		t.Fatalf("restore off: %+v %v", got, err)
	}
}

func TestNextDocTitle(t *testing.T) {
	dir := t.TempDir()
	if n := nextDocTitle(dir); n != "Document" {
		t.Fatalf("got %q", n)
	}
	if _, err := CreateDiskFile(dir, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDiskFile(dir, "", ""); err != nil {
		t.Fatal(err)
	}
	first, err := ReadDiskFile(dir, "Document.md")
	if err != nil {
		t.Fatal(err)
	}
	_ = first
	if _, err := ReadDiskFile(dir, "Document 2.md"); err != nil {
		t.Fatal(err)
	}
}
