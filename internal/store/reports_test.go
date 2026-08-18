package store

import "testing"

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
	if err != nil || len(list) != 1 {
		t.Fatalf("%+v %v", list, err)
	}
}
