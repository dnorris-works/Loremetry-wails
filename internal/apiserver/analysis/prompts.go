package analysis

import "strings"

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
	switch step {
	case "summarize_role":
		b.WriteString("Summarize the following source material for a later craft analysis. Preserve character names, plot beats, themes, and continuity facts. Analysis: ")
	case "summarize_chunk":
		b.WriteString("Summarize this manuscript section for a later full-book analysis. Note POV, pacing, open loops, and craft issues. Analysis: ")
	case "merge_outline":
		b.WriteString("Merge these section summaries into one structured outline (acts, themes, issues, continuity notes). Analysis: ")
	default:
		b.WriteString("Run analysis: ")
	}
	b.WriteString(analysisID)
	b.WriteString("\n\n")
	for _, s := range sources {
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

type Profile int

const (
	ProfileSingle Profile = iota
	ProfileTwoStep
	ProfileChunked
)

func ProfileFor(analysisID string) Profile {
	switch analysisID {
	case "analysis", "wide_analysis", "review_mining", "ai_beta_reader", "competition_report":
		return ProfileChunked
	case "thematic_throughline", "continuity_check", "cross_book_setup_payoff", "series_pacing_comparator", "recurring_motif_theme_series":
		return ProfileTwoStep
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
	default:
		return 1
	}
}
