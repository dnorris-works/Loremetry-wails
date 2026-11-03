package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	seedChapterNotesRules = " Fill only the required markdown headings below (use exactly those heading texts). Use short quotes or location cues. Write \"none\" under a heading if empty. Do not write the full-book report yet—chapter notes only."
	seedFinalMergeRules   = " Merge the chapter notes below into one author-facing markdown report for this analysis. Group patterns across chapters, cite chapter names, rank severity where useful, and do not invent manuscript details not supported by the notes."
)

type catalogSeed struct {
	id                 string
	needs              string
	chapterInstruction string
	chapterHeadings    []string
	finalInstruction   string
}

func seedCatalogPrompts(conn *sql.DB) error {
	const sourceRules = "Treat characters, locations, and themes sources as authoritative for names, relationships, settings, and motifs. Do not invent details that contradict those sources or are unsupported by the manuscript."
	if _, err := conn.Exec(`INSERT OR IGNORE INTO app_settings (key, value) VALUES ('analysis_source_rules', ?)`, sourceRules); err != nil {
		return err
	}

	for _, e := range catalogSeeds() {
		headingsJSON, err := json.Marshal(e.chapterHeadings)
		if err != nil {
			return fmt.Errorf("marshal headings for %s: %w", e.id, err)
		}
		if e.needs != "" {
			_, err = conn.Exec(`
				UPDATE analysis_catalog
				SET needs = ?, chapter_instruction = ?, chapter_headings = ?, final_instruction = ?
				WHERE id = ?`,
				e.needs, e.chapterInstruction, string(headingsJSON), e.finalInstruction, e.id)
		} else {
			_, err = conn.Exec(`
				UPDATE analysis_catalog
				SET chapter_instruction = ?, chapter_headings = ?, final_instruction = ?
				WHERE id = ?`,
				e.chapterInstruction, string(headingsJSON), e.finalInstruction, e.id)
		}
		if err != nil {
			return fmt.Errorf("seed catalog %s: %w", e.id, err)
		}
	}
	if err := seedSourceInjection(conn); err != nil {
		return err
	}
	if err := seedLocalRunners(conn); err != nil {
		return err
	}
	return nil
}

func seedSourceInjection(conn *sql.DB) error {
	rows, err := conn.Query(`SELECT id, needs FROM analysis_catalog`)
	if err != nil {
		return err
	}
	type seedRow struct {
		id    string
		needs string
	}
	var items []seedRow
	for rows.Next() {
		var r seedRow
		if err := rows.Scan(&r.id, &r.needs); err != nil {
			_ = rows.Close()
			return err
		}
		items = append(items, r)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, r := range items {
		mode := "all"
		if stringsContainsStoryElementNeed(r.needs) {
			mode = "selective"
		}
		if _, err := conn.Exec(`UPDATE analysis_catalog SET source_injection = ? WHERE id = ?`, mode, r.id); err != nil {
			return err
		}
	}
	return nil
}

func stringsContainsStoryElementNeed(needsJSON string) bool {
	for _, role := range []string{"characters", "locations", "themes"} {
		if strings.Contains(needsJSON, `"`+role+`"`) {
			return true
		}
	}
	return false
}

func seedLocalRunners(conn *sql.DB) error {
	localIDs := []string{
		"print_production", "line_polish", "vellum_prep", "zeigarnik_analysis",
		"readability_score", "sentence_length_variation", "chapter_balance",
		"dialogue_ratio", "dialogue_tag_audit", "passive_voice", "sticky_sentences",
		"repeated_phrases", "paragraph_length", "opening_closing",
	}
	for _, id := range localIDs {
		if _, err := conn.Exec(`UPDATE analysis_catalog SET local_runner = ? WHERE id = ? AND uses_ai = 0`, id, id); err != nil {
			return err
		}
	}
	configs := map[string]string{
		"line_polish": `{"filter_words":["just","really","very","suddenly","somewhat","actually","basically","literally","quite","rather"],"echo_window":12}`,
		"sticky_sentences": `{"glue_threshold":45}`,
		"zeigarnik_analysis": `{"open_loop_cues":["would","until","tomorrow","later","unless"]}`,
		"print_production": `{"words_per_page":250}`,
	}
	for id, cfg := range configs {
		if _, err := conn.Exec(`UPDATE analysis_catalog SET local_config = ? WHERE id = ?`, cfg, id); err != nil {
			return err
		}
	}
	return nil
}

func catalogSeeds() []catalogSeed {
	marketing := catalogSeed{
		chapterInstruction: "Extract marketable genre and content signals from this chapter (not a full metadata sheet)." + seedChapterNotesRules,
		chapterHeadings:    []string{"Tone and tropes", "Audience signals", "Heat or content notes", "Keyword phrases"},
		finalInstruction:   "Produce the requested KDP/wide marketing analysis from the chapter signal notes and any other sources." + seedFinalMergeRules,
	}
	marketingIDs := []string{
		"analysis", "genre_analysis", "genre_ranking", "kdp_categories", "kdp_keywords",
		"mi_search_terms", "keyword_search", "competition_report", "review_mining",
		"author_analysis", "wide_analysis", "bisac_classification", "discovery_keywords",
		"google_keyword_search",
	}
	out := []catalogSeed{
		{
			id: "chapter_summaries",
			chapterInstruction: "Write a real plot summary of this single chapter for the author. Cover who is present, what happens (in order), stakes or conflict, and any open loops or unanswered questions. Use concrete names and events—no vague marketing language. When character profiles are provided, treat them as the source of truth for names and relationships; do not invent kinship, roles, or labels that contradict the profiles or are not supported by the chapter. Do not write the full-book document yet.",
			chapterHeadings:    []string{"Who is present", "What happens", "Stakes or conflict", "Open loops"},
			finalInstruction:   "Assemble an ordered chapter-by-chapter plot summary document from the chapter notes below. One section per chapter with a clear heading. Keep each summary concrete and faithful to the notes and character profiles; do not invent relationships.",
		},
		{
			id:                 "show_dont_tell",
			chapterInstruction: "Find telling vs showing in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Telling spots", "Strong showing", "Priority fixes"},
			finalInstruction:   "Produce a Show Don't Tell report." + seedFinalMergeRules,
		},
		{
			id:                 "ai_isms",
			chapterInstruction: "Flag prose that reads machine-generated or generic AI voice in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"AI-sounding lines", "Voice strengths", "Priority fixes"},
			finalInstruction:   "Produce an AI-isms report." + seedFinalMergeRules,
		},
		{
			id:                 "story_beat_placement",
			chapterInstruction: "Note story beats and structural milestones in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Beats present", "Timing notes", "Missing or early/late beats"},
			finalInstruction:   "Produce a Story Beat Placement report against common frameworks." + seedFinalMergeRules,
		},
		{
			id:                 "scene_sequel_balance",
			chapterInstruction: "Judge action vs reflection balance in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Scene (action) moments", "Sequel (reflective) moments", "Momentum stalls"},
			finalInstruction:   "Produce a Scene/Sequel Balance report." + seedFinalMergeRules,
		},
		{
			id:                 "pov_discipline",
			needs:              `["manuscript","characters"]`,
			chapterInstruction: "Check POV discipline in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"POV", "Head-hops", "Info leaks", "Severity"},
			finalInstruction:   "Produce a POV Discipline report." + seedFinalMergeRules,
		},
		{
			id:                 "continuity_check",
			chapterInstruction: "Extract facts and flag contradictions or continuity risks in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Facts established", "Possible contradictions", "Open questions"},
			finalInstruction:   "Produce a Continuity Check report." + seedFinalMergeRules,
		},
		{
			id:                 "chekhovs_gun",
			chapterInstruction: "Track setups and payoffs in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Setups planted", "Payoffs hinted", "Unresolved"},
			finalInstruction:   "Produce a Chekhov's Gun report." + seedFinalMergeRules,
		},
		{
			id:                 "red_herring_vs_abandoned",
			chapterInstruction: "Separate intentional misdirection from dropped threads in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Possible red herrings", "Dropped or abandoned threads", "Notes"},
			finalInstruction:   "Produce a Red Herring vs Abandoned report." + seedFinalMergeRules,
		},
		{
			id:                 "foreshadowing_twist_fairness",
			chapterInstruction: "Judge foreshadowing and twist fairness clues in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Foreshadowing planted", "Twist signals", "Fairness notes"},
			finalInstruction:   "Produce a Foreshadowing & Twist Fairness report." + seedFinalMergeRules,
		},
		{
			id:                 "macguffin_clarity",
			chapterInstruction: "Track the driving object or goal in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Goal or object visibility", "Motivation", "Confusion risk"},
			finalInstruction:   "Produce a MacGuffin Clarity report." + seedFinalMergeRules,
		},
		{
			id:                 "timeline_flashback",
			needs:              `["manuscript","characters","locations"]`,
			chapterInstruction: "Track timeline shifts and flashbacks in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Timeline position", "Flashbacks or time jumps", "Clarity issues"},
			finalInstruction:   "Produce a Timeline / Flashback report." + seedFinalMergeRules,
		},
		{
			id:                 "want_vs_need",
			chapterInstruction: "Track external want vs internal need for major characters in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"External wants", "Internal needs", "Alignment or conflict"},
			finalInstruction:   "Produce a Want vs Need report." + seedFinalMergeRules,
		},
		{
			id:                 "thematic_throughline",
			needs:              `["manuscript","themes"]`,
			chapterInstruction: "Note theme signals and thematic drift in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Theme signals", "Supporting scenes", "Drift or contradiction"},
			finalInstruction:   "Produce a Thematic Throughline report." + seedFinalMergeRules,
		},
		{
			id:                 "mirror_foil_character",
			needs:              `["manuscript","characters","themes"]`,
			chapterInstruction: "Note mirror/foil pairings and thematic contrast in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Pairings observed", "Contrast or reflection", "Thematic work"},
			finalInstruction:   "Produce a Mirror/Foil Characters report." + seedFinalMergeRules,
		},
		{
			id:                 "dramatic_irony",
			needs:              `["manuscript","characters"]`,
			chapterInstruction: "Find reader-knows-more moments in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Irony moments", "Tension or humor payoff", "Missed opportunities"},
			finalInstruction:   "Produce a Dramatic Irony report." + seedFinalMergeRules,
		},
		{
			id:                 "stakes_escalation",
			chapterInstruction: "Track how stakes move in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Stakes level", "Escalation or plateau", "Reversals"},
			finalInstruction:   "Produce a Stakes Escalation report." + seedFinalMergeRules,
		},
		{
			id:                 "series_pacing_comparator",
			chapterInstruction: "Note pacing density for series comparison in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Pace feel", "Beat density", "Drag or rush"},
			finalInstruction:   "Produce a Series Pacing Comparator report." + seedFinalMergeRules,
		},
		{
			id:                 "ai_beta_reader",
			needs:              `["manuscript","characters","locations"]`,
			chapterInstruction: "React as a beta reader to this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Engagement", "Confusion", "Put-down risk", "Hook to next"},
			finalInstruction:   "Produce an AI Beta Reader report." + seedFinalMergeRules,
		},
		{
			id:                 "cliffhanger_score",
			chapterInstruction: "Score how hard this chapter ending pulls into the next." + seedChapterNotesRules,
			chapterHeadings:    []string{"Ending beat", "Pull forward (1-5)", "Why"},
			finalInstruction:   "Produce a Cliffhanger Score report." + seedFinalMergeRules,
		},
		{
			id:                 "pacing_curve",
			chapterInstruction: "Score pace and drag risk for this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Pace score (1-5)", "Drag risk", "Notes"},
			finalInstruction:   "Produce a Pacing Curve report." + seedFinalMergeRules,
		},
		{
			id:                 "blurb_builder",
			needs:              `["manuscript","blurb","characters","themes"]`,
			chapterInstruction: "Extract blurb-worthy hooks and stakes from this chapter (no full blurb yet)." + seedChapterNotesRules,
			chapterHeadings:    []string{"Hook moments", "Stakes language", "Spoilers to avoid"},
			finalInstruction:   "Produce blurb variants from the chapter notes and any blurb or theme sources." + seedFinalMergeRules,
		},
		{
			id:                 "content_maturity_advisory",
			chapterInstruction: "Note content maturity signals in this chapter." + seedChapterNotesRules,
			chapterHeadings:    []string{"Heat or intimacy", "Violence or dark content", "Other warnings"},
			finalInstruction:   "Produce a Content & Maturity Advisory." + seedFinalMergeRules,
		},
		{
			id:                 "recurring_motif_theme_series",
			needs:              `["bible","themes"]`,
			finalInstruction:   "Produce a Recurring Motif/Theme (Series) report from the source summaries." + seedFinalMergeRules,
		},
		{
			id:               "cross_book_setup_payoff",
			finalInstruction: "Produce a Cross-Book Setup/Payoff report from the source summaries." + seedFinalMergeRules,
		},
	}
	for _, id := range marketingIDs {
		e := marketing
		e.id = id
		out = append(out, e)
	}
	return out
}
