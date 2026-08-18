package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLocalPrintProduction(t *testing.T) {
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
	if !strings.Contains(out, "Word count") {
		t.Fatalf("got %s", out)
	}
	out, err = RunLocalAnalysis("line_polish", root)
	if err != nil || !strings.Contains(out, "Line-level polish") {
		t.Fatalf("%s %v", out, err)
	}
	if _, err := RunLocalAnalysis("genre_analysis", root); err == nil {
		t.Fatal("expected AI analyses to refuse local run")
	}
}
