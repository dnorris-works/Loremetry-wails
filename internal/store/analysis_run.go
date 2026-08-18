package store

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

func RunLocalAnalysis(id, projectPath string) (string, error) {
	detail, ok := GetAnalysisDetail(id)
	if !ok {
		return "", fmt.Errorf("unknown analysis")
	}
	if detail.UsesAI {
		return "", fmt.Errorf("this analysis uses AI")
	}
	root := ResolveProjectRoot(projectPath)
	if root == "" {
		return "", fmt.Errorf("select a book or series first")
	}
	src := MatchAnalysisSources(root)
	for _, need := range detail.Needs {
		role := src.Role(need)
		if !role.Present || len(role.Files) == 0 {
			return "", fmt.Errorf("missing source: %s", need)
		}
	}
	blobs := CollectNeededText(root, detail.Needs)
	text := joinRoleText(blobs, "manuscript")
	switch id {
	case "print_production":
		return runPrintProduction(text), nil
	case "line_polish":
		return runLinePolish(text), nil
	case "vellum_prep":
		return runVellumPrep(blobs), nil
	case "zeigarnik_analysis":
		return runZeigarnik(blobs), nil
	default:
		return "", fmt.Errorf("this analysis cannot run yet")
	}
}

func joinRoleText(blobs []RoleText, role string) string {
	var b strings.Builder
	for _, blob := range blobs {
		if role != "" && blob.Role != role {
			continue
		}
		b.WriteString(blob.Text)
		b.WriteByte('\n')
	}
	if b.Len() == 0 {
		for _, blob := range blobs {
			b.WriteString(blob.Text)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func runPrintProduction(text string) string {
	words := len(strings.Fields(text))
	chars := len([]rune(text))
	pages6x9 := words / 250
	if pages6x9 < 1 && words > 0 {
		pages6x9 = 1
	}
	spine := float64(pages6x9) * 0.002252
	return fmt.Sprintf(`# Print production

Local estimate from the manuscript. No AI.

- Word count: %d
- Characters: %d
- Approx. 6×9 pages (250 wpp): %d
- Approx. cream spine (in): %.3f

Check KDP print specs for trim, bleed, and cover template before upload.
`, words, chars, pages6x9, spine)
}

var (
	filterWords = regexp.MustCompile(`(?i)\b(just|really|very|suddenly|somewhat|actually|basically|literally|quite|rather)\b`)
)

const echoWindow = 12

func runLinePolish(text string) string {
	words := strings.Fields(text)
	filterHits := filterWords.FindAllString(text, 80)
	lyHits := 0
	for _, w := range words {
		clean := strings.Trim(strings.ToLower(w), ".,;:!?\"'")
		if strings.HasSuffix(clean, "ly") && len(clean) > 4 && clean != "family" && clean != "only" {
			lyHits++
		}
	}
	echoes := findEchoes(words)
	var b strings.Builder
	fmt.Fprintf(&b, "# Line-level polish\n\nHeuristic scan. No AI.\n\n- Words: %d\n- Filter-word hits (sample): %d\n- -ly adverbs: %d\n- Nearby echoes: %d\n\n",
		len(words), len(filterHits), lyHits, len(echoes))
	if len(filterHits) > 0 {
		b.WriteString("## Filter words\n\n")
		seen := map[string]int{}
		for _, h := range filterHits {
			seen[strings.ToLower(h)]++
		}
		for w, n := range seen {
			fmt.Fprintf(&b, "- %s (%d)\n", w, n)
		}
		b.WriteByte('\n')
	}
	if len(echoes) > 0 {
		b.WriteString("## Echoes\n\n")
		limit := 20
		if len(echoes) < limit {
			limit = len(echoes)
		}
		for _, e := range echoes[:limit] {
			fmt.Fprintf(&b, "- %s\n", e)
		}
	}
	return b.String()
}

func findEchoes(words []string) []string {
	var out []string
	norm := make([]string, len(words))
	for i, w := range words {
		norm[i] = strings.ToLower(strings.TrimFunc(w, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		}))
	}
	for i, w := range norm {
		if len(w) < 5 {
			continue
		}
		end := i + echoWindow
		if end > len(norm) {
			end = len(norm)
		}
		for j := i + 1; j < end; j++ {
			if norm[j] == w {
				out = append(out, w)
				break
			}
		}
		if len(out) >= 40 {
			break
		}
	}
	return out
}

func runVellumPrep(blobs []RoleText) string {
	var b strings.Builder
	b.WriteString("# Vellum & Atticus prep\n\nCleaned manuscript markdown. No AI.\n\n")
	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		title := strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name))
		fmt.Fprintf(&b, "# %s\n\n%s\n\n", title, strings.TrimSpace(blob.Text))
	}
	return b.String()
}

func runZeigarnik(blobs []RoleText) string {
	var b strings.Builder
	b.WriteString("# Zeigarnik effect\n\nHeuristic scan for open loops at chapter ends. No AI.\n\n")
	n := 0
	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		end := strings.TrimSpace(blob.Text)
		if end == "" {
			continue
		}
		if len(end) > 400 {
			end = end[len(end)-400:]
		}
		lines := strings.Split(strings.TrimSpace(end), "\n")
		last := strings.TrimSpace(lines[len(lines)-1])
		open := strings.Contains(last, "?") || strings.HasSuffix(last, "...") || strings.HasSuffix(last, "…") ||
			regexp.MustCompile(`(?i)\b(would|until|tomorrow|later|unless)\b`).MatchString(last)
		if open {
			n++
			fmt.Fprintf(&b, "- **%s** — possible open loop: %q\n", blob.Name, last)
		}
	}
	if n == 0 {
		b.WriteString("No strong open-loop chapter endings found.\n")
	}
	return b.String()
}
