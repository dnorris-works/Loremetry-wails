package store

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type AnalysisRunResult struct {
	Markdown string
	Original string
	Proposed string
	DataJSON string
}

func RunLocalAnalysis(id, projectPath string) (AnalysisRunResult, error) {
	detail, ok := GetAnalysisDetail(id)
	if !ok {
		return AnalysisRunResult{}, fmt.Errorf("unknown analysis")
	}
	if detail.UsesAI {
		return AnalysisRunResult{}, fmt.Errorf("this analysis uses AI")
	}
	root := ResolveProjectRoot(projectPath)
	if root == "" {
		return AnalysisRunResult{}, fmt.Errorf("select a book or series first")
	}
	if err := ValidateAnalysisNeeds(root, detail.Needs); err != nil {
		return AnalysisRunResult{}, err
	}
	blobs := CollectNeededText(root, detail.Needs)
	outcome, err := RunLocalFromCatalog(detail, blobs)
	if err != nil {
		return AnalysisRunResult{}, err
	}
	content := outcome.Content
	original := outcome.Original
	proposed := outcome.Proposed
	dataJSON := outcome.DataJSON
	md := ""
	if !detail.UsesMerge {
		md = FormatReportBody(detail.Label, detail.ID, content)
	}
	return AnalysisRunResult{
		Markdown: md,
		Original: original,
		Proposed: proposed,
		DataJSON: dataJSON,
	}, nil
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

func runPrintProduction(text string, wordsPerPage int) string {
	words := len(strings.Fields(text))
	chars := len([]rune(text))
	if wordsPerPage <= 0 {
		wordsPerPage = 250
	}
	pages6x9 := words / wordsPerPage
	if pages6x9 < 1 && words > 0 {
		pages6x9 = 1
	}
	spine := float64(pages6x9) * 0.002252
	return fmt.Sprintf(`- Word count: %d
- Characters: %d
- Approx. 6×9 pages (%d wpp): %d
- Approx. cream spine (in): %.3f

Check KDP print specs for trim, bleed, and cover template before upload.
`, words, chars, wordsPerPage, pages6x9, spine)
}

func defaultFilterWords() []string {
	return []string{"just", "really", "very", "suddenly", "somewhat", "actually", "basically", "literally", "quite", "rather"}
}

func runLinePolish(text string, cfg LocalConfig) string {
	words := strings.Fields(text)
	filterRE := filterWordsRE(cfg.FilterWords)
	filterHits := filterRE.FindAllString(text, 80)
	lyHits := 0
	for _, w := range words {
		clean := strings.Trim(strings.ToLower(w), ".,;:!?\"'")
		if strings.HasSuffix(clean, "ly") && len(clean) > 4 && clean != "family" && clean != "only" {
			lyHits++
		}
	}
	echoes := findEchoes(words, cfg.EchoWindow)
	var b strings.Builder
	fmt.Fprintf(&b, "- Words: %d\n- Filter-word hits (sample): %d\n- -ly adverbs: %d\n- Nearby echoes: %d\n\n",
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

func filterWordsRE(words []string) *regexp.Regexp {
	if len(words) == 0 {
		words = defaultFilterWords()
	}
	parts := make([]string, len(words))
	for i, w := range words {
		parts[i] = regexp.QuoteMeta(w)
	}
	return regexp.MustCompile(`(?i)\b(` + strings.Join(parts, "|") + `)\b`)
}

func findEchoes(words []string, window int) []string {
	if window <= 0 {
		window = 12
	}
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
		end := i + window
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
	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		title := strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name))
		fmt.Fprintf(&b, "# %s\n\n%s\n\n", title, strings.TrimSpace(blob.Text))
	}
	return b.String()
}

var openLoopCue = regexp.MustCompile(`(?i)\b(would|until|tomorrow|later|unless)\b`)

func openLoopCueRE(cfg LocalConfig) *regexp.Regexp {
	if len(cfg.OpenLoopCues) == 0 {
		return openLoopCue
	}
	parts := make([]string, len(cfg.OpenLoopCues))
	for i, c := range cfg.OpenLoopCues {
		parts[i] = regexp.QuoteMeta(c)
	}
	return regexp.MustCompile(`(?i)\b(` + strings.Join(parts, "|") + `)\b`)
}

func chapterEndingOpenLoop(text string, cue *regexp.Regexp) (last string, open bool) {
	end := strings.TrimSpace(text)
	if end == "" {
		return "", false
	}
	sample := end
	if len(sample) > 400 {
		sample = sample[len(sample)-400:]
	}
	lines := strings.Split(strings.TrimSpace(sample), "\n")
	last = strings.TrimSpace(lines[len(lines)-1])
	open = strings.Contains(last, "?") || strings.HasSuffix(last, "...") || strings.HasSuffix(last, "…") ||
		cue.MatchString(last)
	return last, open
}

// runZeigarnik builds manuscript original/proposed for Compare only (no separate report body).
func runZeigarnik(blobs []RoleText, cfg LocalConfig) (original, proposed, dataJSON string) {
	cue := openLoopCueRE(cfg)
	var orig, prop strings.Builder
	type loopHit struct {
		File   string `json:"file"`
		Title  string `json:"title"`
		Ending string `json:"ending"`
	}
	var loops []loopHit
	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		title := strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name))
		text := strings.TrimSpace(blob.Text)
		chapter := fmt.Sprintf("# %s\n\n%s\n\n", title, text)
		orig.WriteString(chapter)

		last, open := chapterEndingOpenLoop(text, cue)
		if open {
			loops = append(loops, loopHit{File: blob.Name, Title: title, Ending: last})
			fmt.Fprintf(&prop, "# %s\n\n%s\n\n> **Open loop** — possible unresolved thread at chapter end.\n\n", title, text)
		} else {
			prop.WriteString(chapter)
		}
	}
	payload, err := json.Marshal(map[string]any{"open_loops": loops})
	if err != nil {
		dataJSON = `{"open_loops":[]}`
	} else {
		dataJSON = string(payload)
	}
	return orig.String(), prop.String(), dataJSON
}
