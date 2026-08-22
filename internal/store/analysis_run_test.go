package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLocalPrintProduction(t *testing.T) {
	ensureCatalogDB(t)
	root := t.TempDir()
	ch := filepath.Join(root, "01_Manuscript", "01_Chapters")
	if err := os.MkdirAll(ch, 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Repeat("word ", 300)
	if err := os.WriteFile(filepath.Join(ch, "Ch-001.md"), []byte(body+"What happens next?"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := RunLocalAnalysis("print_production", root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Markdown, "About this report") || !strings.Contains(out.Markdown, "How to use") {
		t.Fatalf("missing header: %s", out.Markdown)
	}
	if !strings.Contains(out.Markdown, "Word count") {
		t.Fatalf("got %s", out.Markdown)
	}
	out, err = RunLocalAnalysis("line_polish", root)
	if err != nil || !strings.Contains(out.Markdown, "Line-level Polish") {
		t.Fatalf("%s %v", out.Markdown, err)
	}
	if !strings.Contains(out.Markdown, "Filter-word hits") {
		t.Fatalf("missing polish content: %s", out.Markdown)
	}
	if _, err := RunLocalAnalysis("genre_analysis", root); err == nil {
		t.Fatal("expected AI analyses to refuse local run")
	}
}

func TestRunZeigarnikMerge(t *testing.T) {
	ensureCatalogDB(t)
	root := t.TempDir()
	ch := filepath.Join(root, "01_Manuscript", "01_Chapters")
	if err := os.MkdirAll(ch, 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Repeat("word ", 50) + "What happens next?"
	if err := os.WriteFile(filepath.Join(ch, "Ch-001.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := RunLocalAnalysis("zeigarnik_analysis", root)
	if err != nil {
		t.Fatal(err)
	}
	if out.Original == "" || out.Proposed == "" {
		t.Fatalf("expected merge sides, original=%d proposed=%d", len(out.Original), len(out.Proposed))
	}
	if !strings.Contains(out.Markdown, "Open loop") && !strings.Contains(out.Proposed, "Open loop") {
		t.Fatalf("expected open-loop marker in proposed: %s", out.Proposed)
	}
	if strings.TrimSpace(out.Markdown) != "" {
		t.Fatalf("compare-only analysis must not build a report body, got %q", out.Markdown)
	}
}

func TestRunLocalAnalysisMissingSource(t *testing.T) {
	ensureCatalogDB(t)
	root := t.TempDir()
	ch := filepath.Join(root, "01_Manuscript", "01_Chapters")
	if err := os.MkdirAll(ch, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ch, "Ch-001.md"), []byte("Hello world."), 0o644); err != nil {
		t.Fatal(err)
	}
	detail, ok := GetAnalysisDetail("chapter_summaries")
	if !ok {
		t.Fatal("chapter_summaries missing from catalog")
	}
	err := ValidateAnalysisNeeds(root, detail.Needs)
	if err == nil {
		t.Fatal("expected missing characters source")
	}
	if !strings.Contains(err.Error(), "missing source: characters") {
		t.Fatalf("got %v", err)
	}
}