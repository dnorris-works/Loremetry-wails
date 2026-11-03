package analysis

import (
	"loremetry/internal/store"
	"strings"
)

type Source struct {
	Role string `json:"role"`
	Rel  string `json:"rel"`
	Name string `json:"name"`
	Text string `json:"text"`
}

func SystemPrompt(analysisID string) string {
	return "You are Loremetry, an expert fiction editor and literary analyst. Write a clear, grounded markdown report for a novelist. Quote or reference the manuscript where helpful. Avoid generic AI phrasing. Analysis id: " + analysisID
}

func UserPrompt(analysisID string, sources []Source) string {
	return StepPrompt("final", analysisID, sources, nil)
}

func StepPrompt(step, analysisID string, sources []Source, prior []string) string {
	var b strings.Builder
	spec := ChapterNoteSpecFor(analysisID)
	switch step {
	case "summarize_role":
		b.WriteString("Summarize the following source material for a later craft analysis. Preserve character names, plot beats, themes, and continuity facts. Analysis: ")
		b.WriteString(analysisID)
	case "summarize_chunk":
		b.WriteString("Summarize this manuscript section for a later full-book analysis. Note POV, pacing, open loops, and craft issues. Analysis: ")
		b.WriteString(analysisID)
	case "merge_outline":
		b.WriteString("Merge these section summaries into one structured outline (acts, themes, issues, continuity notes). Analysis: ")
		b.WriteString(analysisID)
	case "analyze_chapter":
		b.WriteString(spec.ChapterInstruction)
		b.WriteString(" Analysis: ")
		b.WriteString(analysisID)
		b.WriteString(FormatChapterHeadings(spec.RequiredHeadings))
		if hasStoryElementSource(sources) {
			if rules := store.AnalysisSourceRules(); rules != "" {
				b.WriteString("\n\n")
				b.WriteString(rules)
			}
		}
	default:
		b.WriteString(spec.FinalInstruction)
		b.WriteString(" Analysis: ")
		b.WriteString(analysisID)
		if hasStoryElementSource(sources) {
			if rules := store.AnalysisSourceRules(); rules != "" {
				b.WriteString("\n\n")
				b.WriteString(rules)
			}
		}
	}
	b.WriteString("\n\n")
	for _, s := range orderSourcesForStep(step, sources) {
		b.WriteString("## ")
		b.WriteString(s.Role)
		if s.Name != "" {
			b.WriteString(" / ")
			b.WriteString(s.Name)
		}
		b.WriteString("\n\n")
		b.WriteString(s.Text)
		b.WriteString("\n\n")
	}
	for i, p := range prior {
		b.WriteString("## Prior step ")
		b.WriteString(string(rune('A' + i)))
		b.WriteString("\n\n")
		b.WriteString(p)
		b.WriteString("\n\n")
	}
	return b.String()
}

// orderSourcesForStep puts shared (non-manuscript) sources before the chapter text
// so analyze_chapter calls share a stable prompt prefix for KV / provider caching.
func orderSourcesForStep(step string, sources []Source) []Source {
	if step != "analyze_chapter" || len(sources) <= 1 {
		return sources
	}
	var shared, manuscript []Source
	for _, s := range sources {
		if s.Role == "manuscript" {
			manuscript = append(manuscript, s)
		} else {
			shared = append(shared, s)
		}
	}
	return append(shared, manuscript...)
}

type Profile int

const (
	ProfileSingle Profile = iota
	ProfileTwoStep
	ProfileChunked
	ProfileByChapter
)

func ProfileFor(analysisID string) Profile {
	switch store.AIProfile(analysisID) {
	case "chunked":
		return ProfileChunked
	case "two_step":
		return ProfileTwoStep
	case "by_chapter":
		return ProfileByChapter
	default:
		return ProfileSingle
	}
}

func StepTotal(p Profile) int {
	switch p {
	case ProfileTwoStep:
		return 2
	case ProfileChunked:
		return 3
	case ProfileByChapter:
		return 2
	default:
		return 1
	}
}
