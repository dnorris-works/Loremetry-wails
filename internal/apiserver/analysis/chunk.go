package analysis

import (
	"regexp"
	"strings"
)

var chapterHeadingRE = regexp.MustCompile(`(?m)^(?:#+\s*)?(?:Chapter|CHAPTER)\s+\d+`)

func SplitManuscript(text string, maxChars int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if maxChars <= 0 {
		maxChars = 12000
	}
	if len(text) <= maxChars {
		return []string{text}
	}
	idxs := chapterHeadingRE.FindAllStringIndex(text, -1)
	if len(idxs) == 0 {
		return splitBySize(text, maxChars)
	}
	var chunks []string
	start := 0
	for i, loc := range idxs {
		if i == 0 && loc[0] > 0 {
			continue
		}
		if loc[0] > start {
			part := strings.TrimSpace(text[start:loc[0]])
			if part != "" {
				chunks = append(chunks, part)
			}
			start = loc[0]
		}
	}
	if tail := strings.TrimSpace(text[start:]); tail != "" {
		chunks = append(chunks, tail)
	}
	if len(chunks) == 0 {
		return splitBySize(text, maxChars)
	}
	var out []string
	for _, c := range chunks {
		if len(c) <= maxChars {
			out = append(out, c)
			continue
		}
		out = append(out, splitBySize(c, maxChars)...)
	}
	return out
}

func splitBySize(text string, maxChars int) []string {
	var out []string
	for len(text) > maxChars {
		cut := maxChars
		if sp := strings.LastIndex(text[:cut], "\n\n"); sp > maxChars/2 {
			cut = sp
		}
		out = append(out, strings.TrimSpace(text[:cut]))
		text = strings.TrimSpace(text[cut:])
	}
	if strings.TrimSpace(text) != "" {
		out = append(out, strings.TrimSpace(text))
	}
	return out
}

func TruncateExcerpt(text string, maxChars int) string {
	text = strings.TrimSpace(text)
	if maxChars <= 0 {
		maxChars = 12000
	}
	if len(text) <= maxChars {
		return text
	}
	head := text[:maxChars]
	var b strings.Builder
	b.WriteString(head)
	b.WriteString("\n\n[… manuscript truncated for analysis; full text processed in chunks …]\n")
	for _, loc := range chapterHeadingRE.FindAllStringIndex(text, -1) {
		if loc[0] >= maxChars {
			b.WriteString(text[loc[0]:])
			if idx := strings.Index(b.String()[maxChars:], "\n"); idx > 0 && len(b.String()) > maxChars+200 {
				break
			}
		}
	}
	return b.String()
}
