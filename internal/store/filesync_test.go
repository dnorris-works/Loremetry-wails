package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChapterFileName(t *testing.T) {
	if got := chapterFileName(3, "The Door.md"); got != "03-The Door.md" {
		t.Fatalf("got %s", got)
	}
	if matchKey("03-The Door.md") != "the door" {
		t.Fatalf("match %s", matchKey("03-The Door.md"))
	}
	if stripChapterPrefix("12_intro.txt") != "intro" {
		t.Fatalf("strip")
	}
}

func TestSyncChapterFolder(t *testing.T) {
	s := testStore(t)
	story, err := s.CreateStory(StoryInput{Name: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "02-later.md"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "01-first.md"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetFolderLink(FolderLinkInput{Scope: "story", OwnerID: story.ID, Kind: "chapter", Path: dir}); err != nil {
		t.Fatal(err)
	}
	chs, err := s.ListChapters(story.ID)
	if err != nil || len(chs) != 2 {
		t.Fatalf("chapters %+v %v", chs, err)
	}
	if chs[0].FileName != "01-first.md" || chs[1].FileName != "02-later.md" {
		t.Fatalf("order %+v", chs)
	}
	if err := os.Remove(filepath.Join(dir, "01-first.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "01-inserted.md"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SyncFolder("story", story.ID, "chapter"); err != nil {
		t.Fatal(err)
	}
	chs, err = s.ListChapters(story.ID)
	if err != nil || len(chs) != 2 {
		t.Fatalf("after %+v %v", chs, err)
	}
	if chs[0].FileName != "01-inserted.md" || chs[1].FileName != "02-later.md" {
		t.Fatalf("renumber %+v", chs)
	}
}

func TestSyncNestedChaptersStayNested(t *testing.T) {
	s := testStore(t)
	story, err := s.CreateStory(StoryInput{Name: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	act1 := filepath.Join(dir, "act-1")
	act2 := filepath.Join(dir, "act-2")
	if err := os.MkdirAll(act1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(act2, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(act2, "01-later.md"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(act1, "03-open.md"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetFolderLink(FolderLinkInput{Scope: "story", OwnerID: story.ID, Kind: "chapter", Path: dir}); err != nil {
		t.Fatal(err)
	}
	chs, err := s.ListChapters(story.ID)
	if err != nil || len(chs) != 2 {
		t.Fatalf("chapters %+v %v", chs, err)
	}
	if chs[0].FileName != "01-later.md" || chs[1].FileName != "02-open.md" {
		t.Fatalf("global names %+v", chs)
	}
	if _, err := os.Stat(filepath.Join(act2, "01-later.md")); err != nil {
		t.Fatalf("later should stay in act-2: %v", err)
	}
	if _, err := os.Stat(filepath.Join(act1, "02-open.md")); err != nil {
		t.Fatalf("open should stay in act-1 as 02: %v", err)
	}
	if _, err := os.Stat(filepath.Join(act1, "03-open.md")); !os.IsNotExist(err) {
		t.Fatal("old 03-open.md should be renamed in place")
	}
	var rel1, rel2 string
	if err := s.DB.QueryRow(`SELECT source_rel FROM story_documents WHERE file_name = ?`, "01-later.md").Scan(&rel1); err != nil {
		t.Fatal(err)
	}
	if err := s.DB.QueryRow(`SELECT source_rel FROM story_documents WHERE file_name = ?`, "02-open.md").Scan(&rel2); err != nil {
		t.Fatal(err)
	}
	if rel1 != "act-2/01-later.md" || rel2 != "act-1/02-open.md" {
		t.Fatalf("source_rel %s %s", rel1, rel2)
	}
}
