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
	if _, err := ApplyBookTemplate(filepath.Join(series, "Books"), "Dawn", "Rift", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyBookTemplate(pen, "Lone", "", ""); err != nil {
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
	if p.Series[0].Tree.Folders[0].Label == "" {
		t.Fatal("expected folder label with count")
	}
	if len(p.Books[0].Tree.Folders) == 0 {
		t.Fatal("expected book folder tree")
	}
	if _, err := CreateSeriesOnDisk(root, "Nobody", "", "X"); err == nil {
		t.Fatal("expected missing author error")
	}
}

func TestFolderCountIncludesNestedActs(t *testing.T) {
	root := t.TempDir()
	compiled := filepath.Join(root, "01_Compiled")
	if err := os.MkdirAll(filepath.Join(compiled, "Act-1-Hiding"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(compiled, "Act-1-Hiding", "a.md"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(compiled, "Act-1-Hiding", "b.md"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	node := scanDirNode(compiled)
	if node.Count != 2 || node.Label != "01_Compiled (2)" {
		t.Fatalf("got count %d label %q", node.Count, node.Label)
	}
}

func TestCharacterTypeFolderOrder(t *testing.T) {
	root := t.TempDir()
	chars := filepath.Join(root, "Characters")
	// Create out of order so alpha sort would put Minor before Supporting.
	for _, name := range []string{"Minor", "Zebra", "Main", "Supporting"} {
		if err := os.MkdirAll(filepath.Join(chars, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	node := scanDirNode(chars)
	got := make([]string, 0, len(node.Folders))
	for _, f := range node.Folders {
		got = append(got, f.Name)
	}
	want := []string{"Main", "Supporting", "Minor", "Zebra"}
	if len(got) != len(want) {
		t.Fatalf("folders=%v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("folders=%v want %v", got, want)
		}
	}
}

func TestEnsureStoryElementsThemesFolder(t *testing.T) {
	root := t.TempDir()
	story := filepath.Join(root, "02_Story-Elements")
	if err := os.MkdirAll(story, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(story, "Themes-and-Motifs.md"), []byte("motifs"), 0o644); err != nil {
		t.Fatal(err)
	}
	node := scanDirNode(story)
	if !hasFolder(node, "Themes") {
		t.Fatalf("missing Themes folder: %+v", node.Folders)
	}
	if _, err := os.Stat(filepath.Join(story, "Themes", "Themes-and-Motifs.md")); err != nil {
		t.Fatalf("legacy file should move into Themes: %v", err)
	}
	if _, err := os.Stat(filepath.Join(story, "Themes-and-Motifs.md")); !os.IsNotExist(err) {
		t.Fatal("legacy root file should be gone")
	}
}
