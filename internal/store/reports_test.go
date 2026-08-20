package store

import (
	"strings"
	"testing"
)

func TestFormatReportBody(t *testing.T) {
	got := FormatReportBody("Line-level Polish", "line_polish", "- Words: 10\n")
	if !strings.Contains(got, "# Line-level Polish") {
		t.Fatalf("missing title: %s", got)
	}
	if !strings.Contains(got, "## About this report") || !strings.Contains(got, "## How to use") {
		t.Fatalf("missing sections: %s", got)
	}
	if !strings.Contains(got, "Filter words") && !strings.Contains(got, analysisDescription("line_polish")) {
		t.Fatalf("missing about text: %s", got)
	}
	if !strings.Contains(got, "---") || !strings.Contains(got, "- Words: 10") {
		t.Fatalf("missing content: %s", got)
	}
}

func TestSaveListReports(t *testing.T) {
	s := testStore(t)
	got, err := s.SaveAnalysisReport(AnalysisReport{
		AnalysisID:    "line_polish",
		AnalysisLabel: "Line-level Polish",
		ProjectPath:   "/tmp/book",
		UsesAI:        false,
		Body:          "# ok",
	})
	if err != nil || got.ID == 0 || !strings.HasPrefix(strings.TrimSpace(got.Body), "#") {
		t.Fatalf("%+v %v", got, err)
	}
	list, err := s.ListAnalysisReports()
	if err != nil || len(list) != 1 || list[0].BodySize == 0 {
		t.Fatalf("%+v %v", list, err)
	}
}

func TestSaveReportWithMerge(t *testing.T) {
	s := testStore(t)
	got, err := s.SaveAnalysisReport(AnalysisReport{
		AnalysisID:    "zeigarnik_analysis",
		AnalysisLabel: "Zeigarnik Effect",
		ProjectPath:   "/tmp/book",
		Body:          "# report",
		Original:      "# Ch1\n\nHello.",
		Proposed:      "# Ch1\n\nHello.\n\n> **Open loop**",
	})
	if err != nil {
		t.Fatal(err)
	}
	full, err := s.GetAnalysisReport(got.ID)
	if err != nil {
		t.Fatal(err)
	}
	if full.Original == "" || full.Proposed == "" || !strings.Contains(full.Proposed, "Open loop") {
		t.Fatalf("merge sides not saved: original=%q proposed=%q", full.Original, full.Proposed)
	}
}

func TestLargeReportReturnsFullBody(t *testing.T) {
	s := testStore(t)
	body := strings.Repeat("word ", 20000)
	got, err := s.SaveAnalysisReport(AnalysisReport{
		AnalysisID:    "vellum_prep",
		AnalysisLabel: "Vellum & Atticus Prep",
		ProjectPath:   "/tmp/book",
		Body:          body,
	})
	if err != nil {
		t.Fatal(err)
	}
	full, err := s.GetAnalysisReport(got.ID)
	if err != nil {
		t.Fatal(err)
	}
	if full.Body == "" || full.BodySize <= 0 {
		t.Fatalf("expected full markdown body, got size=%d", full.BodySize)
	}
}
