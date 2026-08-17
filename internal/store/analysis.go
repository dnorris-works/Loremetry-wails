package store

type AnalysisItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type AnalysisGroup struct {
	ID    string          `json:"id"`
	Label string          `json:"label"`
	Items []AnalysisItem  `json:"items"`
}

func AnalysisCatalog() []AnalysisGroup {
	item := func(id, label string) AnalysisItem {
		return AnalysisItem{ID: id, Label: label}
	}
	return []AnalysisGroup{
		{ID: "kdp-wide", Label: "KDP / Wide", Items: []AnalysisItem{
			item("analysis", "KDP Analysis"),
			item("genre_analysis", "Genre Analysis"),
			item("genre_ranking", "Genre Ranking"),
			item("kdp_categories", "KDP Categories"),
			item("kdp_keywords", "KDP Keywords"),
			item("mi_search_terms", "Search Terms"),
			item("keyword_search", "Keyword Search Results"),
			item("competition_report", "Competition Analysis"),
			item("review_mining", "Reader Review Intelligence"),
			item("author_analysis", "Competitor Author Analysis"),
			item("wide_analysis", "Wide Analysis"),
			item("bisac_classification", "BISAC Classification"),
			item("discovery_keywords", "Discovery Keywords"),
			item("google_keyword_search", "Google Keyword Search"),
			item("content_maturity_advisory", "Content & Maturity Advisory"),
			item("wide_metadata_paste", "Wide Metadata Paste Sheet"),
		}},
		{ID: "prose", Label: "Craft — Prose", Items: []AnalysisItem{
			item("show_dont_tell", "Show Don't Tell"),
			item("ai_isms", "AI-isms"),
		}},
		{ID: "structure", Label: "Craft — Structure & Pacing", Items: []AnalysisItem{
			item("story_beat_placement", "Story Beat Placement"),
			item("scene_sequel_balance", "Scene/Sequel Balance"),
			item("pov_discipline", "POV Discipline"),
		}},
		{ID: "plot", Label: "Craft — Plot & Continuity", Items: []AnalysisItem{
			item("continuity_check", "Continuity Check"),
			item("chekhovs_gun", "Chekhov's Gun"),
			item("red_herring_vs_abandoned", "Red Herring vs Abandoned"),
			item("foreshadowing_twist_fairness", "Foreshadowing & Twist Fairness"),
			item("macguffin_clarity", "MacGuffin Clarity"),
			item("timeline_flashback", "Timeline / Flashback"),
		}},
		{ID: "character", Label: "Craft — Character & Theme", Items: []AnalysisItem{
			item("want_vs_need", "Want vs Need"),
			item("thematic_throughline", "Thematic Throughline"),
			item("mirror_foil_character", "Mirror/Foil Characters"),
		}},
		{ID: "engagement", Label: "Craft — Reader Engagement", Items: []AnalysisItem{
			item("zeigarnik_analysis", "Zeigarnik Effect"),
			item("dramatic_irony", "Dramatic Irony"),
			item("stakes_escalation", "Stakes Escalation"),
		}},
		{ID: "series", Label: "Craft — Series", Items: []AnalysisItem{
			item("cross_book_setup_payoff", "Cross-Book Setup/Payoff"),
			item("series_pacing_comparator", "Series Pacing Comparator"),
			item("recurring_motif_theme_series", "Recurring Motif/Theme (Series)"),
		}},
		{ID: "publish", Label: "Publish", Items: []AnalysisItem{
			item("blurb_builder", "Blurb Builder"),
			item("print_production", "Print Production"),
			item("ai_beta_reader", "AI Beta Reader"),
			item("cliffhanger_score", "Cliffhanger Score"),
			item("hook_strength", "Hook Strength"),
			item("pacing_curve", "Pacing Curve"),
			item("line_polish", "Line-level Polish"),
			item("vellum_prep", "Vellum & Atticus Prep"),
		}},
	}
}
