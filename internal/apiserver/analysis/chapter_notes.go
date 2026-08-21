package analysis

import "strings"

// ChapterNoteSpec tells the model how to fill one chapter call and how to synthesize the final report.
type ChapterNoteSpec struct {
	ChapterInstruction string
	RequiredHeadings   []string
	FinalInstruction   string
}

func spec(chapter string, headings []string, final string) ChapterNoteSpec {
	return ChapterNoteSpec{
		ChapterInstruction: chapter,
		RequiredHeadings:   headings,
		FinalInstruction:   final,
	}
}

const chapterNotesRules = " Fill only the required markdown headings below (use exactly those heading texts). Use short quotes or location cues. Write \"none\" under a heading if empty. Do not write the full-book report yet—chapter notes only."

const finalMergeRules = " Merge the chapter notes below into one author-facing markdown report for this analysis. Group patterns across chapters, cite chapter names, rank severity where useful, and do not invent manuscript details not supported by the notes."

// ChapterNoteSpecFor returns structured note instructions for an analysis id.
func ChapterNoteSpecFor(analysisID string) ChapterNoteSpec {
	switch analysisID {
	case "chapter_summaries":
		return spec(
			"Write a real plot summary of this single chapter for the author. Cover who is present, what happens (in order), stakes or conflict, and any open loops or unanswered questions. Use concrete names and events—no vague marketing language. Do not write the full-book document yet.",
			[]string{"Who is present", "What happens", "Stakes or conflict", "Open loops"},
			"Assemble an ordered chapter-by-chapter plot summary document from the chapter notes below. One section per chapter with a clear heading. Keep each summary concrete and faithful to the notes.",
		)

	// Craft — Prose
	case "show_dont_tell":
		return spec(
			"Find telling vs showing in this chapter."+chapterNotesRules,
			[]string{"Telling spots", "Strong showing", "Priority fixes"},
			"Produce a Show Don't Tell report."+finalMergeRules,
		)
	case "ai_isms":
		return spec(
			"Flag prose that reads machine-generated or generic AI voice in this chapter."+chapterNotesRules,
			[]string{"AI-sounding lines", "Voice strengths", "Priority fixes"},
			"Produce an AI-isms report."+finalMergeRules,
		)

	// Craft — Structure & Pacing
	case "story_beat_placement":
		return spec(
			"Note story beats and structural milestones in this chapter."+chapterNotesRules,
			[]string{"Beats present", "Timing notes", "Missing or early/late beats"},
			"Produce a Story Beat Placement report against common frameworks."+finalMergeRules,
		)
	case "scene_sequel_balance":
		return spec(
			"Judge action vs reflection balance in this chapter."+chapterNotesRules,
			[]string{"Scene (action) moments", "Sequel (reflective) moments", "Momentum stalls"},
			"Produce a Scene/Sequel Balance report."+finalMergeRules,
		)
	case "pov_discipline":
		return spec(
			"Check POV discipline in this chapter."+chapterNotesRules,
			[]string{"POV", "Head-hops", "Info leaks", "Severity"},
			"Produce a POV Discipline report."+finalMergeRules,
		)

	// Craft — Plot & Continuity
	case "continuity_check":
		return spec(
			"Extract facts and flag contradictions or continuity risks in this chapter."+chapterNotesRules,
			[]string{"Facts established", "Possible contradictions", "Open questions"},
			"Produce a Continuity Check report."+finalMergeRules,
		)
	case "chekhovs_gun":
		return spec(
			"Track setups and payoffs in this chapter."+chapterNotesRules,
			[]string{"Setups planted", "Payoffs hinted", "Unresolved"},
			"Produce a Chekhov's Gun report."+finalMergeRules,
		)
	case "red_herring_vs_abandoned":
		return spec(
			"Separate intentional misdirection from dropped threads in this chapter."+chapterNotesRules,
			[]string{"Possible red herrings", "Dropped or abandoned threads", "Notes"},
			"Produce a Red Herring vs Abandoned report."+finalMergeRules,
		)
	case "foreshadowing_twist_fairness":
		return spec(
			"Judge foreshadowing and twist fairness clues in this chapter."+chapterNotesRules,
			[]string{"Foreshadowing planted", "Twist signals", "Fairness notes"},
			"Produce a Foreshadowing & Twist Fairness report."+finalMergeRules,
		)
	case "macguffin_clarity":
		return spec(
			"Track the driving object or goal in this chapter."+chapterNotesRules,
			[]string{"Goal or object visibility", "Motivation", "Confusion risk"},
			"Produce a MacGuffin Clarity report."+finalMergeRules,
		)
	case "timeline_flashback":
		return spec(
			"Track timeline shifts and flashbacks in this chapter."+chapterNotesRules,
			[]string{"Timeline position", "Flashbacks or time jumps", "Clarity issues"},
			"Produce a Timeline / Flashback report."+finalMergeRules,
		)

	// Craft — Character & Theme
	case "want_vs_need":
		return spec(
			"Track external want vs internal need for major characters in this chapter."+chapterNotesRules,
			[]string{"External wants", "Internal needs", "Alignment or conflict"},
			"Produce a Want vs Need report."+finalMergeRules,
		)
	case "thematic_throughline":
		return spec(
			"Note theme signals and thematic drift in this chapter."+chapterNotesRules,
			[]string{"Theme signals", "Supporting scenes", "Drift or contradiction"},
			"Produce a Thematic Throughline report."+finalMergeRules,
		)
	case "mirror_foil_character":
		return spec(
			"Note mirror/foil pairings and thematic contrast in this chapter."+chapterNotesRules,
			[]string{"Pairings observed", "Contrast or reflection", "Thematic work"},
			"Produce a Mirror/Foil Characters report."+finalMergeRules,
		)

	// Craft — Reader Engagement / Publish craft
	case "dramatic_irony":
		return spec(
			"Find reader-knows-more moments in this chapter."+chapterNotesRules,
			[]string{"Irony moments", "Tension or humor payoff", "Missed opportunities"},
			"Produce a Dramatic Irony report."+finalMergeRules,
		)
	case "stakes_escalation":
		return spec(
			"Track how stakes move in this chapter."+chapterNotesRules,
			[]string{"Stakes level", "Escalation or plateau", "Reversals"},
			"Produce a Stakes Escalation report."+finalMergeRules,
		)
	case "series_pacing_comparator":
		return spec(
			"Note pacing density for series comparison in this chapter."+chapterNotesRules,
			[]string{"Pace feel", "Beat density", "Drag or rush"},
			"Produce a Series Pacing Comparator report."+finalMergeRules,
		)
	case "ai_beta_reader":
		return spec(
			"React as a beta reader to this chapter."+chapterNotesRules,
			[]string{"Engagement", "Confusion", "Put-down risk", "Hook to next"},
			"Produce an AI Beta Reader report."+finalMergeRules,
		)
	case "cliffhanger_score":
		return spec(
			"Score how hard this chapter ending pulls into the next."+chapterNotesRules,
			[]string{"Ending beat", "Pull forward (1-5)", "Why"},
			"Produce a Cliffhanger Score report."+finalMergeRules,
		)
	case "pacing_curve":
		return spec(
			"Score pace and drag risk for this chapter."+chapterNotesRules,
			[]string{"Pace score (1-5)", "Drag risk", "Notes"},
			"Produce a Pacing Curve report."+finalMergeRules,
		)
	case "blurb_builder":
		return spec(
			"Extract blurb-worthy hooks and stakes from this chapter (no full blurb yet)."+chapterNotesRules,
			[]string{"Hook moments", "Stakes language", "Spoilers to avoid"},
			"Produce blurb variants from the chapter notes and any bible/blurb sources."+finalMergeRules,
		)
	case "content_maturity_advisory":
		return spec(
			"Note content maturity signals in this chapter."+chapterNotesRules,
			[]string{"Heat or intimacy", "Violence or dark content", "Other warnings"},
			"Produce a Content & Maturity Advisory."+finalMergeRules,
		)

	// Light marketing / KDP signals (still one call per chapter)
	case "analysis", "genre_analysis", "genre_ranking", "kdp_categories", "kdp_keywords",
		"mi_search_terms", "keyword_search", "competition_report", "review_mining",
		"author_analysis", "wide_analysis", "bisac_classification", "discovery_keywords",
		"google_keyword_search":
		return marketingChapterSpec()

	default:
		return spec(
			"Analyze this single chapter for the named craft analysis. List concrete findings with short quotes or location cues. Do not write the full book report yet—chapter notes only.",
			[]string{"Findings", "Quotes or cues", "Severity"},
			"Produce the analysis report from the chapter notes below."+finalMergeRules,
		)
	}
}

func marketingChapterSpec() ChapterNoteSpec {
	return spec(
		"Extract marketable genre and content signals from this chapter (not a full metadata sheet)."+chapterNotesRules,
		[]string{"Tone and tropes", "Audience signals", "Heat or content notes", "Keyword phrases"},
		"Produce the requested KDP/wide marketing analysis from the chapter signal notes and any other sources."+finalMergeRules,
	)
}

// FormatChapterHeadings lists required markdown headings for the chapter prompt.
func FormatChapterHeadings(headings []string) string {
	if len(headings) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\nRequired headings (use exactly):\n")
	for _, h := range headings {
		b.WriteString("## ")
		b.WriteString(h)
		b.WriteString("\n")
	}
	return b.String()
}
