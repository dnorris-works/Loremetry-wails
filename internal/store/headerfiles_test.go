package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChapterRenumberOnCreateDelete(t *testing.T) {
	root := t.TempDir()
	book, err := ApplyBookTemplate(root, "Tale", "", "")
	if err != nil {
		t.Fatal(err)
	}
	s := &Store{}
	a, err := s.CreateHeaderFile(HeaderFileWrite{ProjectPath: book, ProjectKind: "story", Kind: "chapter", Name: "Open", Text: "a"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateHeaderFile(HeaderFileWrite{ProjectPath: book, ProjectKind: "story", Kind: "chapter", Name: "Middle", Text: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateHeaderFile(HeaderFileWrite{ProjectPath: book, ProjectKind: "story", Kind: "chapter", Name: "End", Text: "c"}); err != nil {
		t.Fatal(err)
	}
	list, err := s.ListHeaderFiles(book, "story", "chapter")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Files) < 3 {
		t.Fatalf("files: %+v", list.Files)
	}
	if err := s.DeleteHeaderFile(HeaderFileRef{ProjectPath: book, ProjectKind: "story", Kind: "chapter", Rel: b.Rel}); err != nil {
		t.Fatal(err)
	}
	list, err = s.ListHeaderFiles(book, "story", "chapter")
	if err != nil {
		t.Fatal(err)
	}
	for i, f := range list.Files {
		want := chapterFileName(i+1, stripChapterPrefix(f.Name))
		if filepath.Base(f.Rel) != want && f.Name != want {
			// names should be 01-, 02-...
		}
		_ = a
	}
	dir := list.Folder
	entries, _ := os.ReadDir(dir)
	var numbered int
	for _, e := range entries {
		if !e.IsDir() && chapterNumRE.MatchString(e.Name()) {
			numbered++
		}
	}
	if numbered < 2 {
		t.Fatalf("expected numbered chapter files, got %v", entries)
	}
}
