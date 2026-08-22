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
	if !ok || !genre.UsesAI {
		t.Fatalf("genre %+v %v", genre, ok)
	}
	if len(genre.DependsOn) != 0 {
		t.Fatalf("genre should not depend on chapter_summaries, got %+v", genre.DependsOn)
	}
	summary, ok := GetAnalysisDetail("chapter_summaries")
	if !ok || !summary.UsesAI {
		t.Fatalf("chapter_summaries %+v %v", summary, ok)
	}
	if len(summary.Needs) != 2 || summary.Needs[0] != "manuscript" || summary.Needs[1] != "characters" {
		t.Fatalf("chapter_summaries needs %+v", summary.Needs)
	}
	if summary.ChapterInstruction == "" || len(summary.ChapterHeadings) != 4 {
		t.Fatalf("chapter_summaries prompts %+v", summary)
	}
	pov, ok := GetAnalysisDetail("pov_discipline")
	if !ok || len(pov.Needs) != 2 || pov.Needs[1] != "characters" {
		t.Fatalf("pov_discipline needs %+v", pov.Needs)
	}
	throughline, ok := GetAnalysisDetail("thematic_throughline")
	if !ok || len(throughline.Needs) != 2 || throughline.Needs[1] != "themes" {
		t.Fatalf("thematic_throughline needs %+v", throughline.Needs)
	}
	blurb, ok := GetAnalysisDetail("blurb_builder")
	if !ok || len(blurb.Needs) != 4 {
		t.Fatalf("blurb_builder needs %+v", blurb.Needs)
	}
}
