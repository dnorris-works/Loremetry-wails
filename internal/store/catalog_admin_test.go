package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateAnalysisCatalog(t *testing.T) {
	ensureCatalogDB(t)
	got, err := UpdateAnalysisCatalog(CatalogUpdateInput{
		ID:              "pov_discipline",
		Label:           "POV Discipline",
		Description:     "POV shifts",
		Usage:           "Fix POV",
		Needs:           []string{"manuscript", "characters"},
		DependsOn:       []string{},
		UsesAI:          true,
		AIProfile:       "by_chapter",
		SourceInjection: "selective",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceInjection != "selective" {
		t.Fatalf("source_injection %q", got.SourceInjection)
	}
}

func TestLocalRunnerFromCatalog(t *testing.T) {
	ensureCatalogDB(t)
	detail, ok := GetAnalysisDetail("line_polish")
	if !ok {
		t.Fatal("missing line_polish")
	}
	if detail.LocalRunner != "line_polish" {
		t.Fatalf("local_runner %q", detail.LocalRunner)
	}
	root := t.TempDir()
	book, err := ApplyBookTemplate(root, "Tale", "", "Act")
	if err != nil {
		t.Fatal(err)
	}
	chDir := filepath.Join(book, "01_Manuscript", "01_Chapters")
	if err := os.WriteFile(filepath.Join(chDir, "001-Scene-Title.md"), []byte("Just really very suddenly she walked."), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := RunLocalAnalysis("line_polish", book)
	if err != nil {
		t.Fatal(err)
	}
	if out.Markdown == "" {
		t.Fatal("empty markdown")
	}
}
