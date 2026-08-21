package store

import "testing"

func TestAnalysisCatalog(t *testing.T) {
	ensureCatalogDB(t)
	groups := AnalysisCatalog()
	if len(groups) != 9 {
		t.Fatalf("groups %d", len(groups))
	}
	n := 0
	for _, g := range groups {
		if g.ID == "" || g.Label == "" || len(g.Items) == 0 {
			t.Fatalf("%+v", g)
		}
		n += len(g.Items)
	}
	if n < 30 {
		t.Fatalf("items %d", n)
	}
	got, ok := GetAnalysisDetail("show_dont_tell")
	if !ok || got.Label == "" || got.Description == "" || !got.UsesAI {
		t.Fatalf("detail %+v %v", got, ok)
	}
	local, ok := GetAnalysisDetail("line_polish")
	if !ok || local.UsesAI || local.UsesMerge {
		t.Fatalf("local %+v %v", local, ok)
	}
	zeig, ok := GetAnalysisDetail("zeigarnik_analysis")
	if !ok || zeig.UsesAI || !zeig.UsesMerge {
		t.Fatalf("zeigarnik %+v %v", zeig, ok)
	}
	genre, ok := GetAnalysisDetail("genre_analysis")
	if !ok || len(genre.DependsOn) == 0 || genre.DependsOn[0] != "chapter_summaries" {
		t.Fatalf("genre deps %+v %v", genre, ok)
	}
}
