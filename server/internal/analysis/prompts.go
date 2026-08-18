package analysis

import "strings"

func SystemPrompt(analysisID string) string {
	return "You are Loremetry. Write a clear analysis report in markdown for a novelist. Analysis id: " + analysisID
}

func UserPrompt(analysisID string, sources []Source) string {
	var b strings.Builder
	b.WriteString("Run analysis: ")
	b.WriteString(analysisID)
	b.WriteString("\n\n")
	for _, s := range sources {
		b.WriteString("## ")
		b.WriteString(s.Role)
		b.WriteString(" / ")
		b.WriteString(s.Name)
		b.WriteString("\n\n")
		b.WriteString(s.Text)
		b.WriteString("\n\n")
	}
	return b.String()
}

type Source struct {
	Role string `json:"role"`
	Rel  string `json:"rel"`
	Name string `json:"name"`
	Text string `json:"text"`
}
