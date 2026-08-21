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
	src := MatchAnalysisSources(root)
	for _, need := range detail.Needs {
		role := src.Role(need)
		if !role.Present || len(role.Files) == 0 {
			return AnalysisRunResult{}, fmt.Errorf("missing source: %s", need)
		}
	}
	blobs := CollectNeededText(root, detail.Needs)
	text := joinRoleText(blobs, "manuscript")
	var content, original, proposed, dataJSON string
	switch id {
	case "print_production":
		content = runPrintProduction(text)
		dataJSON = fmt.Sprintf(`{"words":%d}`, len(strings.Fields(text)))
	case "line_polish":
		content = runLinePolish(text)
		dataJSON = `{"kind":"line_polish"}`
	case "vellum_prep":
		content = runVellumPrep(blobs)
		dataJSON = `{"kind":"vellum_prep"}`
	case "zeigarnik_analysis":
		original, proposed, dataJSON = runZeigarnik(blobs)
	case "readability_score":
		content = runReadabilityScore(blobs)
		dataJSON = `{"kind":"readability_score"}`
	case "sentence_length_variation":
		content = runSentenceLengthVariation(blobs)
		dataJSON = `{"kind":"sentence_length_variation"}`
	case "chapter_balance":
		content = runChapterBalance(blobs)
		dataJSON = `{"kind":"chapter_balance"}`
	case "dialogue_ratio":
		content = runDialogueRatio(blobs)
		dataJSON = `{"kind":"dialogue_ratio"}`
	case "dialogue_tag_audit":
		content = runDialogueTagAudit(blobs)
		dataJSON = `{"kind":"dialogue_tag_audit"}`
	case "passive_voice":
		content = runPassiveVoice(blobs)
		dataJSON = `{"kind":"passive_voice"}`
	case "sticky_sentences":
		content, dataJSON = runStickySentences(blobs)
	case "repeated_phrases":
		content = runRepeatedPhrases(blobs)
		dataJSON = `{"kind":"repeated_phrases"}`
	case "paragraph_length":
		content = runParagraphLength(blobs)
		dataJSON = `{"kind":"paragraph_length"}`
	case "opening_closing":
		content = runOpeningClosing(blobs)
		dataJSON = `{"kind":"opening_closing"}`
	default:
		return AnalysisRunResult{}, fmt.Errorf("this analysis cannot run yet")
	}
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

func runPrintProduction(text string) string {
	words := len(strings.Fields(text))
	chars := len([]rune(text))
	pages6x9 := words / 250
	if pages6x9 < 1 && words > 0 {
		pages6x9 = 1
	}
	spine := float64(pages6x9) * 0.002252
	return fmt.Sprintf(`- Word count: %d
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

func chapterEndingOpenLoop(text string) (last string, open bool) {
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
		openLoopCue.MatchString(last)
	return last, open
}

// runZeigarnik builds manuscript original/proposed for Compare only (no separate report body).
func runZeigarnik(blobs []RoleText) (original, proposed, dataJSON string) {
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

		last, open := chapterEndingOpenLoop(text)
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
