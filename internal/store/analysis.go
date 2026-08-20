package store

import "strings"

type AnalysisItem struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Needs       []string `json:"needs"`
	DependsOn   []string `json:"depends_on"`
	UsesAI      bool     `json:"uses_ai"`
	UsesMerge   bool     `json:"uses_merge"`
}

type AnalysisDetail struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Group       string   `json:"group"`
	Description string   `json:"description"`
	Needs       []string `json:"needs"`
	DependsOn   []string `json:"depends_on"`
	UsesAI      bool     `json:"uses_ai"`
	UsesMerge   bool     `json:"uses_merge"`
}

type AnalysisGroup struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Items []AnalysisItem `json:"items"`
}

func AnalysisCatalog() []AnalysisGroup {
	item := func(id, label string, needs ...string) AnalysisItem {
		if needs == nil {
			needs = []string{}
		}
		return AnalysisItem{
			ID: id, Label: label, Description: analysisDescription(id),
			Needs: needs, DependsOn: analysisDependsOn(id),
			UsesAI: analysisUsesAI(id), UsesMerge: analysisUsesMerge(id),
		}
	}
	return []AnalysisGroup{
		{ID: "kdp-wide", Label: "KDP / Wide", Items: []AnalysisItem{
			item("chapter_summaries", "Chapter Summaries", "manuscript"),
			item("analysis", "KDP Analysis", "manuscript", "plot"),
			item("genre_analysis", "Genre Analysis", "manuscript"),
			item("genre_ranking", "Genre Ranking", "manuscript"),
			item("kdp_categories", "KDP Categories", "manuscript"),
			item("kdp_keywords", "KDP Keywords", "manuscript", "plot"),
			item("mi_search_terms", "Search Terms", "manuscript"),
			item("keyword_search", "Keyword Search Results", "manuscript"),
			item("competition_report", "Competition Analysis", "manuscript"),
			item("review_mining", "Reader Review Intelligence", "manuscript"),
			item("author_analysis", "Competitor Author Analysis", "manuscript"),
			item("wide_analysis", "Wide Analysis", "manuscript", "plot"),
			item("bisac_classification", "BISAC Classification", "manuscript"),
			item("discovery_keywords", "Discovery Keywords", "manuscript"),
			item("google_keyword_search", "Google Keyword Search", "manuscript"),
			item("content_maturity_advisory", "Content & Maturity Advisory", "manuscript"),
			item("wide_metadata_paste", "Wide Metadata Paste Sheet", "plot"),
		}},
		{ID: "prose", Label: "Craft — Prose", Items: []AnalysisItem{
			item("show_dont_tell", "Show Don't Tell", "manuscript"),
			item("ai_isms", "AI-isms", "manuscript"),
		}},
		{ID: "structure", Label: "Craft — Structure & Pacing", Items: []AnalysisItem{
			item("story_beat_placement", "Story Beat Placement", "manuscript"),
			item("scene_sequel_balance", "Scene/Sequel Balance", "manuscript"),
			item("pov_discipline", "POV Discipline", "manuscript"),
		}},
		{ID: "plot", Label: "Craft — Plot & Continuity", Items: []AnalysisItem{
			item("continuity_check", "Continuity Check", "manuscript", "characters", "locations"),
			item("chekhovs_gun", "Chekhov's Gun", "manuscript"),
			item("red_herring_vs_abandoned", "Red Herring vs Abandoned", "manuscript"),
			item("foreshadowing_twist_fairness", "Foreshadowing & Twist Fairness", "manuscript"),
			item("macguffin_clarity", "MacGuffin Clarity", "manuscript"),
			item("timeline_flashback", "Timeline / Flashback", "manuscript"),
		}},
		{ID: "character", Label: "Craft — Character & Theme", Items: []AnalysisItem{
			item("want_vs_need", "Want vs Need", "manuscript", "characters"),
			item("thematic_throughline", "Thematic Throughline", "manuscript", "bible"),
			item("mirror_foil_character", "Mirror/Foil Characters", "manuscript", "characters"),
		}},
		{ID: "engagement", Label: "Craft — Reader Engagement", Items: []AnalysisItem{
			item("zeigarnik_analysis", "Zeigarnik Effect", "manuscript"),
			item("dramatic_irony", "Dramatic Irony", "manuscript"),
			item("stakes_escalation", "Stakes Escalation", "manuscript"),
		}},
		{ID: "series", Label: "Craft — Series", Items: []AnalysisItem{
			item("cross_book_setup_payoff", "Cross-Book Setup/Payoff", "bible", "characters"),
			item("series_pacing_comparator", "Series Pacing Comparator", "manuscript"),
			item("recurring_motif_theme_series", "Recurring Motif/Theme (Series)", "bible"),
		}},
		{ID: "publish", Label: "Publish", Items: []AnalysisItem{
			item("blurb_builder", "Blurb Builder", "manuscript", "bible", "blurb"),
			item("print_production", "Print Production", "manuscript"),
			item("ai_beta_reader", "AI Beta Reader", "manuscript"),
			item("cliffhanger_score", "Cliffhanger Score", "manuscript"),
			item("hook_strength", "Hook Strength", "manuscript"),
			item("pacing_curve", "Pacing Curve", "manuscript"),
			item("line_polish", "Line-level Polish", "manuscript"),
			item("vellum_prep", "Vellum & Atticus Prep", "manuscript"),
		}},
	}
}

func GetAnalysisDetail(id string) (AnalysisDetail, bool) {
	id = strings.TrimSpace(id)
	for _, g := range AnalysisCatalog() {
		for _, it := range g.Items {
			if it.ID == id {
				return AnalysisDetail{
					ID:          it.ID,
					Label:       it.Label,
					Group:       g.Label,
					Description: it.Description,
					Needs:       it.Needs,
					DependsOn:   it.DependsOn,
					UsesAI:      it.UsesAI,
					UsesMerge:   it.UsesMerge,
				}, true
			}
		}
	}
	return AnalysisDetail{}, false
}

// CollectPrerequisites returns transitive depends_on ids in dependency order (deps before dependents).
// The primary analysis id itself is not included.
func CollectPrerequisites(id string) []string {
	id = strings.TrimSpace(id)
	visited := map[string]bool{}
	var out []string
	var walk func(string)
	walk = func(cur string) {
		detail, ok := GetAnalysisDetail(cur)
		if !ok {
			return
		}
		for _, dep := range detail.DependsOn {
			dep = strings.TrimSpace(dep)
			if dep == "" || dep == id || visited[dep] {
				continue
			}
			visited[dep] = true
			walk(dep)
			out = append(out, dep)
		}
	}
	walk(id)
	return out
}

// AnalysisRunQueue returns prerequisites + primary in run order.
func AnalysisRunQueue(id string) []string {
	id = strings.TrimSpace(id)
	deps := CollectPrerequisites(id)
	out := make([]string, 0, len(deps)+1)
	out = append(out, deps...)
	out = append(out, id)
	return out
}

func analysisDependsOn(id string) []string {
	switch id {
	case "genre_analysis", "wide_analysis", "discovery_keywords", "content_maturity_advisory":
		return []string{"chapter_summaries"}
	case "genre_ranking", "kdp_categories", "kdp_keywords", "bisac_classification":
		return []string{"chapter_summaries", "genre_analysis"}
	case "mi_search_terms":
		return []string{"chapter_summaries", "analysis"}
	case "analysis":
		return []string{"chapter_summaries", "mi_search_terms"}
	case "keyword_search":
		return []string{"chapter_summaries", "analysis"}
	case "competition_report", "review_mining", "author_analysis":
		return []string{"mi_search_terms"}
	case "google_keyword_search":
		return []string{"chapter_summaries", "discovery_keywords"}
	case "wide_metadata_paste":
		return []string{"bisac_classification", "discovery_keywords"}
	case "blurb_builder":
		return []string{"chapter_summaries"}
	default:
		return []string{}
	}
}

func analysisUsesAI(id string) bool {
	switch id {
	case "zeigarnik_analysis", "print_production", "line_polish", "vellum_prep":
		return false
	default:
		return true
	}
}

// analysisUsesMerge marks analyses that return original/proposed for the Compare merge view.
func analysisUsesMerge(id string) bool {
	switch id {
	case "zeigarnik_analysis":
		return true
	default:
		return false
	}
}

func analysisDescription(id string) string {
	switch id {
	case "chapter_summaries":
		return "AI genre-signal summary per chapter. Required by most KDP and Wide analyses."
	case "analysis":
		return "Genre, Kindle and paperback categories, print BISAC, seven keywords, and ready-to-paste KDP metadata."
	case "genre_analysis":
		return "Industry genre classification, ranking, comparable titles, and reader demographic."
	case "genre_ranking":
		return "Score the manuscript against known genres independently."
	case "kdp_categories":
		return "Best-fit Amazon browse paths for Kindle and/or paperback."
	case "kdp_keywords":
		return "Seven KDP keyword strings (50 characters or fewer)."
	case "mi_search_terms":
		return "Short Amazon-style phrases for competition research."
	case "keyword_search":
		return "Amazon keyword volume and competition data."
	case "competition_report":
		return "Market landscape: niche competitiveness, comps, pricing, and viability."
	case "review_mining":
		return "What readers love and hate in competitor reviews, plus positioning notes."
	case "author_analysis":
		return "Competitor catalogs: cadence, pricing, series vs standalone, and takeaways."
	case "wide_analysis":
		return "BISAC, discovery keywords, Google SEO, content advisory, and wide paste metadata."
	case "bisac_classification":
		return "BISAC subject codes for Ingram, Apple, Kobo, and aggregators."
	case "discovery_keywords":
		return "Ten wide-store SEO phrases."
	case "google_keyword_search":
		return "Google search volume for wide-store phrases."
	case "content_maturity_advisory":
		return "Heat level, content warnings, and age guidance for wide stores."
	case "wide_metadata_paste":
		return "Copy-ready wide metadata for aggregators and stores."
	case "zeigarnik_analysis":
		return "Heuristic scan for open loops, cliffhangers, and unresolved threads. No AI."
	case "continuity_check":
		return "Finds contradicted facts in a manuscript or across a series."
	case "show_dont_tell":
		return "Flags passages that tell emotion or judgment instead of showing it."
	case "ai_isms":
		return "Flags prose that often reads as machine-generated."
	case "chekhovs_gun":
		return "Early setups (objects, skills, promises) and whether they pay off."
	case "red_herring_vs_abandoned":
		return "Separates intentional misdirection from dropped plot threads."
	case "foreshadowing_twist_fairness":
		return "Whether foreshadowing is fair and twists feel earned."
	case "macguffin_clarity":
		return "Whether the driving object or goal is clear and motivates action."
	case "want_vs_need":
		return "External want vs internal need for major characters."
	case "thematic_throughline":
		return "How the central theme holds across scenes and arcs."
	case "mirror_foil_character":
		return "Reflect and contrast pairings and what they do for theme."
	case "pov_discipline":
		return "POV shifts, head-hopping, and information leaks."
	case "story_beat_placement":
		return "Beat timing against common story frameworks."
	case "scene_sequel_balance":
		return "Action vs reflective passages and where momentum stalls."
	case "timeline_flashback":
		return "Whether timeline shifts and flashbacks clarify or confuse."
	case "dramatic_irony":
		return "Moments the reader knows more than the characters."
	case "stakes_escalation":
		return "How stakes rise, plateau, or reverse across the arc."
	case "cross_book_setup_payoff":
		return "Series setups planted in earlier books that should pay off later."
	case "series_pacing_comparator":
		return "Pacing compared across books in a series."
	case "recurring_motif_theme_series":
		return "Motifs and themes across books — cohesion vs contradiction."
	case "blurb_builder":
		return "Amazon, back-cover, and BookBub description variants."
	case "print_production":
		return "Page count, trim, spine, and print checklists from word count. No AI."
	case "ai_beta_reader":
		return "Chapter-by-chapter reader reactions, engagement, and put-down risk."
	case "cliffhanger_score":
		return "How hard each chapter ending pulls into the next."
	case "hook_strength":
		return "Whether a browsing reader would keep going past page one."
	case "pacing_curve":
		return "Per-chapter pace and drag-risk scores."
	case "line_polish":
		return "Filter words, echoes, adverbs, and rough passives. No AI."
	case "vellum_prep":
		return "Clean markdown manuscript for Vellum or Atticus import. No AI."
	default:
		return ""
	}
}

func analysisUsage(id string) string {
	switch id {
	case "chapter_summaries":
		return "Run this first for KDP/Wide stacks, or select a dependent analysis and it will be included automatically."
	case "analysis":
		return "Copy categories, BISAC, and keywords into KDP. Save this report before you change the manuscript."
	case "genre_analysis":
		return "Use the top genre and comps when choosing categories and writing your blurb."
	case "genre_ranking":
		return "Compare scores across genres; pick the strongest fit for metadata and positioning."
	case "kdp_categories":
		return "Enter the suggested Kindle and paperback browse paths in KDP category fields."
	case "kdp_keywords":
		return "Paste each string into KDP’s seven keyword slots (50 characters or fewer)."
	case "mi_search_terms":
		return "Use these phrases in Amazon search and competition tools to size the niche."
	case "keyword_search":
		return "Favor high-volume, moderate-competition terms when refining KDP keywords."
	case "competition_report":
		return "Weigh niche density and comps before locking genre, price, and series strategy."
	case "review_mining":
		return "Steal reader language for blurbs and fix pain points they complain about in comps."
	case "author_analysis":
		return "Benchmark release cadence, pricing, and series length against nearby competitors."
	case "wide_analysis":
		return "Apply BISAC, discovery keywords, and content notes across wide stores and aggregators."
	case "bisac_classification":
		return "Enter the recommended BISAC codes in Ingram, Apple, Kobo, or your aggregator."
	case "discovery_keywords":
		return "Add these phrases to wide-store keyword and SEO fields."
	case "google_keyword_search":
		return "Prioritize phrases with real search volume when filling wide discovery keywords."
	case "content_maturity_advisory":
		return "Match store maturity settings and content warnings to the advisory below."
	case "wide_metadata_paste":
		return "Copy each block into the matching aggregator or store metadata field."
	case "zeigarnik_analysis":
		return "Use Compare to review flagged chapter endings; strengthen weak open loops and leave intentional closures alone."
	case "continuity_check":
		return "Fix contradicted facts in the manuscript (or series bible) before the next draft pass."
	case "show_dont_tell":
		return "Rewrite flagged telling passages to show emotion through action and sensory detail."
	case "ai_isms":
		return "Revise flagged lines so they sound like your voice, not generic machine prose."
	case "chekhovs_gun":
		return "Pay off unused setups or cut them so early mentions earn their keep."
	case "red_herring_vs_abandoned":
		return "Keep intentional misdirection; restore or remove dropped threads."
	case "foreshadowing_twist_fairness":
		return "Add fair clues where twists feel cheap; trim over-telegraphing where they feel obvious."
	case "macguffin_clarity":
		return "Clarify the driving object or goal if characters (or readers) lose track of why it matters."
	case "want_vs_need":
		return "Align scenes so external want and internal need pull characters through the arc."
	case "thematic_throughline":
		return "Reinforce or prune scenes that drift from the central theme."
	case "mirror_foil_character":
		return "Use mirror/foil pairings to sharpen theme; cut pairings that do no thematic work."
	case "pov_discipline":
		return "Fix head-hops and information leaks so each scene stays in the intended POV."
	case "story_beat_placement":
		return "Move, add, or cut beats so structural milestones land where the framework expects."
	case "scene_sequel_balance":
		return "Add action or reflection where the report shows momentum stalling."
	case "timeline_flashback":
		return "Clarify or cut flashbacks that confuse chronology rather than deepen stakes."
	case "dramatic_irony":
		return "Lean into reader-knows-more moments where tension or humor pays off."
	case "stakes_escalation":
		return "Raise or reset stakes where the arc plateaus; cut reversals that undercut tension."
	case "cross_book_setup_payoff":
		return "Track series setups that still need payoff in a later book or epilogue."
	case "series_pacing_comparator":
		return "Adjust book length and beat density so series pacing feels consistent."
	case "recurring_motif_theme_series":
		return "Keep motifs coherent across books; resolve contradictions in theme or symbol."
	case "blurb_builder":
		return "Pick a variant for Amazon, back cover, or BookBub and paste it into that channel."
	case "print_production":
		return "Use word count and spine estimates when choosing trim and ordering a KDP/Ingram cover template."
	case "ai_beta_reader":
		return "Prioritize chapters with high put-down risk; revise hooks and endings there first."
	case "cliffhanger_score":
		return "Strengthen weak chapter endings so readers turn the page."
	case "hook_strength":
		return "Rewrite the opening if a browsing reader would stop before page two."
	case "pacing_curve":
		return "Trim or split chapters marked as drag risk; keep high-pace chapters intact."
	case "line_polish":
		return "Search the manuscript for listed filter words and echoes; cut or replace the worst repeats."
	case "vellum_prep":
		return "Copy or export this markdown into Vellum or Atticus as a new book import."
	default:
		return "Review the findings below and apply changes in your manuscript or metadata as needed."
	}
}
