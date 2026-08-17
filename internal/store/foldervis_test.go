package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterWritingTreeHidesNamedFolders(t *testing.T) {
	root := t.TempDir()
	pen := filepath.Join(root, "Ada")
	book, err := ApplyBookTemplate(pen, "Tale", "", "")
	if err != nil {
		t.Fatal(err)
	}
	tree := ListWritingTree(root)
	if len(tree.Pens) != 1 || len(tree.Pens[0].Books) != 1 {
		t.Fatalf("tree: %+v", tree)
	}
	vis := FolderVisibility{HiddenNames: []string{"07_Marketing"}, Overrides: map[string]map[string]string{}}
	filtered := FilterWritingTree(tree, vis)
	if hasFolder(filtered.Pens[0].Books[0].Tree, "07_Marketing") {
		t.Fatal("expected marketing hidden")
	}
	if !hasFolder(filtered.Pens[0].Books[0].Tree, "01_Manuscript") {
		t.Fatal("expected manuscript visible")
	}
	ms := findFolder(filtered.Pens[0].Books[0].Tree, "01_Manuscript")
	if ms == nil || ms.Required {
		t.Fatal("01_Manuscript is not an analysis role")
	}
	draft := findFolder(*ms, "01_Chapters")
	if draft == nil || !draft.Required {
		t.Fatal("expected chapters marked required")
	}
	if f := findFolder(filtered.Pens[0].Books[0].Tree, "04_Research"); f != nil && f.Required {
		t.Fatal("research is not needed by any analysis")
	}
	plot := findFolder(filtered.Pens[0].Books[0].Tree, "03_Plot")
	if plot == nil || !plot.Required {
		t.Fatal("expected plot marked required")
	}
	vis.ShowHidden = true
	shown := FilterWritingTree(ListWritingTree(root), vis)
	m := findFolder(shown.Pens[0].Books[0].Tree, "07_Marketing")
	if m == nil || !m.Hidden {
		t.Fatal("expected hidden flag when showing hidden")
	}
	vis.ShowHidden = false
	vis.Overrides = map[string]map[string]string{book: {"07_Marketing": "show"}}
	forced := FilterWritingTree(ListWritingTree(root), vis)
	if !hasFolder(forced.Pens[0].Books[0].Tree, "07_Marketing") {
		t.Fatal("expected per-story show override")
	}
}

func TestTemplateFolderNames(t *testing.T) {
	names := TemplateFolderNames()
	if len(names.Books) == 0 || len(names.Series) == 0 {
		t.Fatalf("%+v", names)
	}
}

func hasFolder(n DirNode, name string) bool {
	return findFolder(n, name) != nil
}

func findFolder(n DirNode, name string) *DirNode {
	for i := range n.Folders {
		if n.Folders[i].Name == name {
			return &n.Folders[i]
		}
		if found := findFolder(n.Folders[i], name); found != nil {
			return found
		}
	}
	return nil
}

func TestFolderVisibilityRoundTrip(t *testing.T) {
	s := testStore(t)
	if _, err := s.SetHiddenFolderNames([]string{"07_Marketing", "07_Marketing", " "}); err != nil {
		t.Fatal(err)
	}
	vis, err := s.GetFolderVisibility()
	if err != nil || len(vis.HiddenNames) != 1 {
		t.Fatalf("%+v %v", vis, err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "book"), 0o755); err != nil {
		t.Fatal(err)
	}
	book := filepath.Join(root, "book")
	if _, err := s.SetFolderOverride(FolderOverrideInput{ProjectPath: book, Key: "00_Admin", Mode: "hide"}); err != nil {
		t.Fatal(err)
	}
	vis, _ = s.GetFolderVisibility()
	if vis.Overrides[book]["00_Admin"] != "hide" {
		t.Fatalf("%+v", vis)
	}
}
