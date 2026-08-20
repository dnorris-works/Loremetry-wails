package store

import (
	"strings"
	"testing"
)

func TestSaveListReports(t *testing.T) {
	s := testStore(t)
	got, err := s.SaveAnalysisReport(AnalysisReport{
		AnalysisID:    "line_polish",
		AnalysisLabel: "Line-level Polish",
		ProjectPath:   "/tmp/book",
		UsesAI:        false,
		Body:          "# ok",
	})
	if err != nil || got.ID == 0 || got.Body != "# ok" {
		t.Fatalf("%+v %v", got, err)
	}
	list, err := s.ListAnalysisReports()
	if err != nil || len(list) != 1 || list[0].BodySize == 0 {
		t.Fatalf("%+v %v", list, err)
	}
}

func TestLargeReportMetaOnly(t *testing.T) {
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
	meta, err := s.GetAnalysisReport(got.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !meta.Large || meta.Body != "" || meta.BodySize <= ReportInlineMax {
		t.Fatalf("expected large meta-only report, got large=%v body=%d size=%d", meta.Large, len(meta.Body), meta.BodySize)
	}
	chunk, err := s.GetAnalysisReportBodyRange(got.ID, 0, 1000)
	if err != nil || chunk.Text == "" || chunk.Total <= ReportInlineMax {
		t.Fatalf("chunk: %+v %v", chunk, err)
	}
}
