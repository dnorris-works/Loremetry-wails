package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListWritingTree(t *testing.T) {
	root := t.TempDir()
	pen := filepath.Join(root, "Ada")
	if _, err := ApplySeriesTemplate(pen, "Rift"); err != nil {
		t.Fatal(err)
	}
	series := filepath.Join(pen, "Rift")
	if _, err := ApplyBookTemplate(filepath.Join(series, "Books"), "Dawn", "Rift"); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyBookTemplate(pen, "Lone", ""); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pen, "LooseNotes"), 0o755); err != nil {
		t.Fatal(err)
	}

	tree := ListWritingTree(root)
	if len(tree.Problems) != 0 {
		t.Fatalf("root problems: %+v", tree.Problems)
	}
	if len(tree.Pens) != 1 || tree.Pens[0].Name != "Ada" {
		t.Fatalf("pens: %+v", tree.Pens)
	}
	p := tree.Pens[0]
	if len(p.Series) != 1 || p.Series[0].Name != "Rift" {
		t.Fatalf("series: %+v", p.Series)
	}
	if len(p.Series[0].Books) != 1 || p.Series[0].Books[0].Name != "Dawn" {
		t.Fatalf("series books: %+v", p.Series[0].Books)
	}
	if len(p.Books) != 1 || p.Books[0].Name != "Lone" {
		t.Fatalf("standalone: %+v", p.Books)
	}
	if len(p.Problems) == 0 {
		t.Fatal("expected problem for LooseNotes")
	}
	if len(p.Series[0].Tree.Folders) == 0 {
		t.Fatal("expected series folder tree")
	}
	if len(p.Books[0].Tree.Folders) == 0 {
		t.Fatal("expected book folder tree")
	}
	if _, err := CreateSeriesOnDisk(root, "Nobody", "", "X"); err == nil {
		t.Fatal("expected missing author error")
	}
}
