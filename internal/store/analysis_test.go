package store

import "testing"

func TestAnalysisCatalog(t *testing.T) {
	groups := AnalysisCatalog()
	if len(groups) != 8 {
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
	if !ok || local.UsesAI {
		t.Fatalf("local %+v %v", local, ok)
	}
}
