package analysis

import (
	"strings"

	"loremetry/internal/store"
)

// ChapterNoteSpec tells the model how to fill one chapter call and how to synthesize the final report.
type ChapterNoteSpec struct {
	ChapterInstruction string
	RequiredHeadings   []string
	FinalInstruction   string
}

func defaultChapterNoteSpec() ChapterNoteSpec {
	return ChapterNoteSpec{
		ChapterInstruction: "Analyze this single chapter for the named craft analysis. List concrete findings with short quotes or location cues. Do not write the full book report yet—chapter notes only.",
		RequiredHeadings:   []string{"Findings", "Quotes or cues", "Severity"},
		FinalInstruction:   "Produce the analysis report from the chapter notes below. Merge the chapter notes below into one author-facing markdown report for this analysis. Group patterns across chapters, cite chapter names, rank severity where useful, and do not invent manuscript details not supported by the notes.",
	}
}

// ChapterNoteSpecFromDetail builds prompt instructions from catalog rows.
func ChapterNoteSpecFromDetail(d store.AnalysisDetail) ChapterNoteSpec {
	if strings.TrimSpace(d.ChapterInstruction) == "" && strings.TrimSpace(d.FinalInstruction) == "" {
		return defaultChapterNoteSpec()
	}
	spec := ChapterNoteSpec{
		ChapterInstruction: d.ChapterInstruction,
		RequiredHeadings:   d.ChapterHeadings,
		FinalInstruction:   d.FinalInstruction,
	}
	if len(spec.RequiredHeadings) == 0 && strings.TrimSpace(spec.ChapterInstruction) != "" {
		spec.RequiredHeadings = defaultChapterNoteSpec().RequiredHeadings
	}
	if strings.TrimSpace(spec.FinalInstruction) == "" {
		spec.FinalInstruction = defaultChapterNoteSpec().FinalInstruction
	}
	return spec
}

// ChapterNoteSpecFor returns structured note instructions for an analysis id.
func ChapterNoteSpecFor(analysisID string) ChapterNoteSpec {
	detail, ok := store.GetAnalysisDetail(analysisID)
	if !ok {
		return defaultChapterNoteSpec()
	}
	return ChapterNoteSpecFromDetail(detail)
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

func hasStoryElementSource(sources []Source) bool {
	for _, s := range sources {
		switch s.Role {
		case "characters", "locations", "themes":
			return true
		}
	}
	return false
}
