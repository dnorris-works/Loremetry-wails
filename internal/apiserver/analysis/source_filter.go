package analysis

import (
	"regexp"
	"strings"
	"unicode"
)

var wordBoundary = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// FilterStoryElementsForChapter keeps non-story-element sources and only profiles
// whose label appears in the chapter text (token optimization).
func FilterStoryElementsForChapter(other []Source, chapter Source) []Source {
	chapterText := strings.ToLower(chapter.Text)
	out := make([]Source, 0, len(other))
	for _, s := range other {
		switch s.Role {
		case "characters", "locations", "themes":
			if sourceRelevantToChapter(s, chapterText) {
				out = append(out, s)
			}
		default:
			out = append(out, s)
		}
	}
	return out
}

func sourceRelevantToChapter(s Source, chapterText string) bool {
	for _, label := range profileLabels(s) {
		if label != "" && containsWord(chapterText, strings.ToLower(label)) {
			return true
		}
	}
	return false
}

func profileLabels(s Source) []string {
	var labels []string
	if name := strings.TrimSuffix(s.Name, ".md"); name != "" && !strings.EqualFold(name, s.Role) {
		labels = append(labels, name)
	}
	first := firstHeadingOrLine(s.Text)
	if first != "" {
		labels = append(labels, first)
	}
	return labels
}

func firstHeadingOrLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "#")
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func containsWord(haystack, word string) bool {
	word = strings.TrimSpace(word)
	if word == "" || haystack == "" {
		return false
	}
	for _, part := range wordBoundary.Split(haystack, -1) {
		if part == word {
			return true
		}
	}
	// Also match if the label is a single capitalized name in prose.
	for _, r := range strings.Fields(haystack) {
		clean := strings.TrimFunc(r, func(ch rune) bool {
			return !unicode.IsLetter(ch)
		})
		if strings.EqualFold(clean, word) {
			return true
		}
	}
	return false
}
