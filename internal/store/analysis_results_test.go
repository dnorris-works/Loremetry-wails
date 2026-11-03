package store

import (
	"strings"
	"testing"
)

func TestUpsertAnalysisResult(t *testing.T) {
	s := testStore(t)
	got, err := s.UpsertAnalysisResult(AnalysisResult{
		AnalysisID:  "zeigarnik_analysis",
		Label:       "Zeigarnik Effect",
		ProjectPath: "/tmp/book",
		DataJSON:    `{"open_loops":[{"file":"a.md"}]}`,
		Markdown:    "# report",
		Original:    "orig",
		Proposed:    "prop",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Original != "orig" || got.Proposed != "prop" || !strings.Contains(got.DataJSON, "open_loops") {
		t.Fatalf("%+v", got)
	}
	again, err := s.UpsertAnalysisResult(AnalysisResult{
		AnalysisID:  "zeigarnik_analysis",
		Label:       "Zeigarnik Effect",
		ProjectPath: "/tmp/book",
		DataJSON:    `{"open_loops":[]}`,
		Markdown:    "# report2",
	})
	if err != nil || again.Markdown != "# report2" {
		t.Fatalf("%+v %v", again, err)
	}
	list, err := s.ListAnalysisResults("/tmp/book")
	if err != nil || len(list) != 1 {
		t.Fatalf("%+v %v", list, err)
	}
}

func TestAnalysisRunQueue(t *testing.T) {
	ensureCatalogDB(t)
	q := AnalysisRunQueue("genre_analysis")
	if len(q) != 1 || q[0] != "genre_analysis" {
		t.Fatalf("%v", q)
	}
	rank := AnalysisRunQueue("genre_ranking")
	if len(rank) < 2 || rank[len(rank)-1] != "genre_ranking" || rank[0] != "genre_analysis" {
		t.Fatalf("%v", rank)
	}
	for _, id := range rank {
		if id == "chapter_summaries" {
			t.Fatalf("queue should not include chapter_summaries: %v", rank)
		}
	}
	z := AnalysisRunQueue("zeigarnik_analysis")
	if len(z) != 1 || z[0] != "zeigarnik_analysis" {
		t.Fatalf("%v", z)
	}
}
