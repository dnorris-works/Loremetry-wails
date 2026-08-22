package store

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// fictionCharsPerLine is the wrap width used for sticky-sentence line numbers
// (~5.5×8 fiction text block).
const fictionCharsPerLine = 60

// sentenceEndRe matches sentence-ending punctuation followed by whitespace or end-of-string.
var sentenceEndRe = regexp.MustCompile(`([.!?])(\s|$)`)

// paragraphSepRe matches double newlines (with optional whitespace between) used to split paragraphs.
var paragraphSepRe = regexp.MustCompile(`\n\s*\n`)

// abbreviations is a set of common abbreviations whose trailing period should
// not be treated as a sentence boundary.
var abbreviations = map[string]bool{
	"mr":   true,
	"mrs":  true,
	"ms":   true,
	"dr":   true,
	"prof": true,
	"sr":   true,
	"jr":   true,
	"st":   true,
	"vs":   true,
	"etc":  true,
	"inc":  true,
	"ltd":  true,
	"gen":  true,
	"sgt":  true,
	"cpl":  true,
	"pvt":  true,
	"rev":  true,
	"hon":  true,
	"capt": true,
	"col":  true,
	"maj":  true,
	"lt":   true,
}

// isVowel returns true if the rune is a vowel (including y).
func isVowel(ch rune) bool {
	return ch == 'a' || ch == 'e' || ch == 'i' || ch == 'o' || ch == 'u' || ch == 'y'
}

// countSyllables estimates the number of syllables in a word using a
// vowel-cluster heuristic with silent-e removal.
func countSyllables(word string) int {
	w := strings.ToLower(strings.TrimSpace(word))
	if w == "" {
		return 0
	}
	// Remove trailing 'e' (silent e) if word is longer than 2 chars
	if len(w) > 2 && w[len(w)-1] == 'e' {
		w = w[:len(w)-1]
	}
	count := 0
	inVowel := false
	for _, ch := range w {
		if isVowel(ch) {
			if !inVowel {
				count++
				inVowel = true
			}
		} else {
			inVowel = false
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

// splitSentences splits text into sentences on sentence-ending punctuation
// (.!?) followed by whitespace or end-of-string. It respects common
// abbreviations (Mr., Dr., etc.) so their periods are not treated as
// sentence boundaries.
func splitSentences(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	// Find all sentence-end positions using regex, then filter out abbreviations.
	var sentences []string
	start := 0

	matches := sentenceEndRe.FindAllStringIndex(text, -1)
	if matches == nil {
		// No sentence-ending punctuation found — return whole text as one sentence.
		return []string{text}
	}

	for _, m := range matches {
		punctIdx := m[0] // index of the punctuation mark

		// Check if this period follows a known abbreviation.
		if text[punctIdx] == '.' && isAbbreviation(text, punctIdx) {
			continue
		}

		sentence := strings.TrimSpace(text[start : punctIdx+1])
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
		// Advance start past the punctuation and any whitespace.
		start = m[1]
	}

	// Capture any trailing text after the last split point.
	if start < len(text) {
		tail := strings.TrimSpace(text[start:])
		if tail != "" {
			sentences = append(sentences, tail)
		}
	}

	return sentences
}

// isAbbreviation checks whether the period at position idx in text follows
// a known abbreviation word.
func isAbbreviation(text string, idx int) bool {
	// Walk backwards from idx to find the start of the word before the period.
	wordEnd := idx
	wordStart := idx - 1
	for wordStart >= 0 && unicode.IsLetter(rune(text[wordStart])) {
		wordStart--
	}
	wordStart++ // move to the first letter

	if wordStart >= wordEnd {
		return false
	}

	word := strings.ToLower(text[wordStart:wordEnd])
	return abbreviations[word]
}

// splitParagraphs splits text into paragraphs on double newlines. Only
// non-empty paragraphs are returned; whitespace is trimmed from each.
func splitParagraphs(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	parts := paragraphSepRe.Split(text, -1)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// glueWords is the set of common function words used to compute sticky-sentence ratios.
var glueWords = map[string]bool{
	"the": true, "is": true, "of": true, "to": true, "in": true,
	"and": true, "a": true, "that": true, "it": true, "for": true,
	"with": true, "was": true, "on": true, "as": true, "are": true,
	"at": true, "be": true, "this": true, "have": true, "from": true,
	"or": true, "an": true, "by": true, "not": true, "but": true,
	"what": true, "we": true, "can": true, "if": true, "has": true,
	"do": true, "no": true, "so": true, "will": true, "they": true,
	"all": true, "its": true, "he": true, "she": true, "his": true,
	"her": true, "my": true, "our": true, "your": true, "their": true,
	"one": true, "had": true, "were": true, "been": true, "would": true,
	"could": true, "should": true, "may": true, "might": true, "shall": true,
	"did": true, "does": true, "than": true, "them": true, "these": true,
	"those": true, "then": true, "just": true, "also": true, "about": true,
	"up": true, "out": true, "into": true, "over": true, "after": true,
	"when": true, "where": true, "how": true, "who": true, "which": true,
}

// isGlueWord returns true if the given word is a glue word (case-insensitive).
func isGlueWord(w string) bool {
	return glueWords[strings.ToLower(w)]
}

// normalizeWord lowercases a word and strips surrounding punctuation.
func normalizeWord(w string) string {
	return strings.ToLower(strings.TrimFunc(w, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}))
}

// runReadabilityScore computes Flesch Reading Ease and Flesch-Kincaid Grade
// Level per chapter and overall, outputting a markdown report with a summary
// section and a per-chapter table.
func runReadabilityScore(blobs []RoleText) string {
	type chapterStats struct {
		name      string
		words     int
		sentences int
		syllables int
		fre       float64
		fkgl      float64
		flagged   bool
	}

	var chapters []chapterStats
	var totalWords, totalSentences, totalSyllables int

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		if text == "" {
			continue
		}

		words := strings.Fields(text)
		wordCount := len(words)
		sents := splitSentences(text)
		sentCount := len(sents)

		if wordCount == 0 || sentCount == 0 {
			continue
		}

		syllableCount := 0
		for _, w := range words {
			syllableCount += countSyllables(w)
		}

		avgSentLen := float64(wordCount) / float64(sentCount)
		avgSyllPerWord := float64(syllableCount) / float64(wordCount)

		fre := 206.835 - 1.015*avgSentLen - 84.6*avgSyllPerWord
		fkgl := 0.39*avgSentLen + 11.8*avgSyllPerWord - 15.59

		// Clamp FRE to [0, 100]
		if fre < 0 {
			fre = 0
		}
		if fre > 100 {
			fre = 100
		}

		chapters = append(chapters, chapterStats{
			name:      strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name)),
			words:     wordCount,
			sentences: sentCount,
			syllables: syllableCount,
			fre:       fre,
			fkgl:      fkgl,
		})

		totalWords += wordCount
		totalSentences += sentCount
		totalSyllables += syllableCount
	}

	if len(chapters) == 0 {
		return "No manuscript chapters with content found."
	}

	// Manuscript-wide aggregates computed from totals (not averages of averages).
	avgSentLen := float64(totalWords) / float64(totalSentences)
	avgSyllPerWord := float64(totalSyllables) / float64(totalWords)
	overallFRE := 206.835 - 1.015*avgSentLen - 84.6*avgSyllPerWord
	overallFKGL := 0.39*avgSentLen + 11.8*avgSyllPerWord - 15.59

	if overallFRE < 0 {
		overallFRE = 0
	}
	if overallFRE > 100 {
		overallFRE = 100
	}

	// Compute mean and standard deviation of per-chapter FRE.
	meanFRE := 0.0
	for _, ch := range chapters {
		meanFRE += ch.fre
	}
	meanFRE /= float64(len(chapters))

	variance := 0.0
	for _, ch := range chapters {
		diff := ch.fre - meanFRE
		variance += diff * diff
	}
	variance /= float64(len(chapters))
	stddev := math.Sqrt(variance)

	// Flag chapters deviating >1 standard deviation from the mean FRE.
	for i := range chapters {
		if math.Abs(chapters[i].fre-meanFRE) > stddev {
			chapters[i].flagged = true
		}
	}

	// Build output.
	var b strings.Builder

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Manuscript FRE:** %.1f\n", overallFRE)
	fmt.Fprintf(&b, "- **Manuscript FKGL:** %.1f\n", overallFKGL)
	fmt.Fprintf(&b, "- **Total words:** %d\n", totalWords)
	fmt.Fprintf(&b, "- **Total sentences:** %d\n", totalSentences)
	fmt.Fprintf(&b, "- **Chapters analyzed:** %d\n", len(chapters))
	b.WriteString("- **Genre fiction reference:** grade 4–8 / FRE 60–80\n")
	b.WriteString("\n")

	// Per-chapter table.
	b.WriteString("## Per-Chapter Breakdown\n\n")
	b.WriteString("| Chapter | Words | Sentences | Syllables | FRE | FKGL | Flag |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")

	for _, ch := range chapters {
		flag := ""
		if ch.flagged {
			flag = "⚠️"
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %.1f | %.1f | %s |\n",
			ch.name, ch.words, ch.sentences, ch.syllables, ch.fre, ch.fkgl, flag)
	}

	return b.String()
}

// runSentenceLengthVariation computes per-chapter sentence length statistics
// (avg, min, max, stddev in words) and detects monotonous passages (5+
// consecutive sentences within ±3 words of each other, i.e. max-min ≤ 6).
// Output is a markdown report with a summary section and per-chapter table.
func runSentenceLengthVariation(blobs []RoleText) string {
	type chapterStats struct {
		name       string
		sentences  int
		avg        float64
		min        int
		max        int
		stddev     float64
		monotonous int
	}

	var chapters []chapterStats
	var totalSentences int
	var allLengths []int
	var totalMonotonous int

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		if text == "" {
			continue
		}

		sents := splitSentences(text)
		if len(sents) == 0 {
			continue
		}

		// Compute sentence lengths in words.
		lengths := make([]int, len(sents))
		sum := 0
		minLen := math.MaxInt
		maxLen := 0
		for i, s := range sents {
			wc := len(strings.Fields(s))
			lengths[i] = wc
			sum += wc
			if wc < minLen {
				minLen = wc
			}
			if wc > maxLen {
				maxLen = wc
			}
		}

		n := len(lengths)
		avg := float64(sum) / float64(n)

		// Standard deviation.
		variance := 0.0
		for _, l := range lengths {
			diff := float64(l) - avg
			variance += diff * diff
		}
		variance /= float64(n)
		sd := math.Sqrt(variance)

		// Detect monotonous passages: 5+ consecutive sentences where
		// max - min within the window is ≤ 6.
		monotonous := 0
		if n >= 5 {
			runStart := 0
			for runStart <= n-5 {
				// Find the extent of a monotonous run starting at runStart.
				// Initialize with a window of 5.
				winMin := lengths[runStart]
				winMax := lengths[runStart]
				for k := runStart + 1; k < runStart+5; k++ {
					if lengths[k] < winMin {
						winMin = lengths[k]
					}
					if lengths[k] > winMax {
						winMax = lengths[k]
					}
				}

				if winMax-winMin > 6 {
					runStart++
					continue
				}

				// We have at least 5 consecutive qualifying sentences.
				// Extend the run as far as possible.
				runEnd := runStart + 5
				for runEnd < n {
					candidate := lengths[runEnd]
					newMin := winMin
					newMax := winMax
					if candidate < newMin {
						newMin = candidate
					}
					if candidate > newMax {
						newMax = candidate
					}
					if newMax-newMin > 6 {
						break
					}
					winMin = newMin
					winMax = newMax
					runEnd++
				}

				monotonous++
				// Skip past this passage to avoid counting overlapping ones.
				runStart = runEnd
			}
		}

		chapters = append(chapters, chapterStats{
			name:       strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name)),
			sentences:  n,
			avg:        avg,
			min:        minLen,
			max:        maxLen,
			stddev:     sd,
			monotonous: monotonous,
		})

		totalSentences += n
		allLengths = append(allLengths, lengths...)
		totalMonotonous += monotonous
	}

	if len(chapters) == 0 {
		return "No manuscript chapters with content found."
	}

	// Manuscript-wide stats.
	overallSum := 0
	overallMin := math.MaxInt
	overallMax := 0
	for _, l := range allLengths {
		overallSum += l
		if l < overallMin {
			overallMin = l
		}
		if l > overallMax {
			overallMax = l
		}
	}
	overallAvg := float64(overallSum) / float64(len(allLengths))

	overallVariance := 0.0
	for _, l := range allLengths {
		diff := float64(l) - overallAvg
		overallVariance += diff * diff
	}
	overallVariance /= float64(len(allLengths))
	overallStddev := math.Sqrt(overallVariance)

	// Build output.
	var b strings.Builder

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Total sentences:** %d\n", totalSentences)
	fmt.Fprintf(&b, "- **Average sentence length:** %.1f words\n", overallAvg)
	fmt.Fprintf(&b, "- **Min sentence length:** %d words\n", overallMin)
	fmt.Fprintf(&b, "- **Max sentence length:** %d words\n", overallMax)
	fmt.Fprintf(&b, "- **Standard deviation:** %.1f\n", overallStddev)
	fmt.Fprintf(&b, "- **Monotonous passages found:** %d\n", totalMonotonous)
	b.WriteString("- **Monotonous:** 5+ consecutive sentences within ±3 words of each other\n")
	b.WriteString("\n")

	// Per-chapter table.
	b.WriteString("## Per-Chapter Breakdown\n\n")
	b.WriteString("| Chapter | Sentences | Avg | Min | Max | StdDev | Monotonous |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")

	for _, ch := range chapters {
		fmt.Fprintf(&b, "| %s | %d | %.1f | %d | %d | %.1f | %d |\n",
			ch.name, ch.sentences, ch.avg, ch.min, ch.max, ch.stddev, ch.monotonous)
	}

	return b.String()
}

// runChapterBalance reports per-chapter word counts with summary statistics
// (mean, median, min, max, coefficient of variation) and flags outlier
// chapters whose word count is >2× or <0.5× the median.
func runChapterBalance(blobs []RoleText) string {
	type chapterEntry struct {
		name  string
		words int
	}

	var chapters []chapterEntry

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		wc := len(strings.Fields(text))
		chapters = append(chapters, chapterEntry{
			name:  strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name)),
			words: wc,
		})
	}

	if len(chapters) == 0 {
		return "No manuscript chapters found."
	}

	// Compute statistics.
	n := len(chapters)
	sum := 0
	minWords := chapters[0].words
	maxWords := chapters[0].words
	for _, ch := range chapters {
		sum += ch.words
		if ch.words < minWords {
			minWords = ch.words
		}
		if ch.words > maxWords {
			maxWords = ch.words
		}
	}
	mean := float64(sum) / float64(n)

	// Median: sort a copy of word counts and pick the middle value.
	sorted := make([]int, n)
	for i, ch := range chapters {
		sorted[i] = ch.words
	}
	sort.Ints(sorted)

	var median float64
	if n%2 == 0 {
		median = float64(sorted[n/2-1]+sorted[n/2]) / 2.0
	} else {
		median = float64(sorted[n/2])
	}

	// Standard deviation and coefficient of variation.
	variance := 0.0
	for _, ch := range chapters {
		diff := float64(ch.words) - mean
		variance += diff * diff
	}
	variance /= float64(n)
	stddev := math.Sqrt(variance)

	cv := 0.0
	if mean > 0 {
		cv = (stddev / mean) * 100
	}

	// Build output.
	var b strings.Builder

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Chapters:** %d\n", n)
	fmt.Fprintf(&b, "- **Mean word count:** %.0f\n", mean)
	fmt.Fprintf(&b, "- **Median word count:** %.0f\n", median)
	fmt.Fprintf(&b, "- **Min word count:** %d\n", minWords)
	fmt.Fprintf(&b, "- **Max word count:** %d\n", maxWords)
	fmt.Fprintf(&b, "- **Coefficient of variation:** %.1f%%\n", cv)
	b.WriteString("- **Outlier threshold:** >2× or <0.5× median\n")
	b.WriteString("\n")

	// Per-chapter table sorted by chapter order (insertion order).
	b.WriteString("## Per-Chapter Word Counts\n\n")
	b.WriteString("| Chapter | Words | Flag |\n")
	b.WriteString("|---|---|---|\n")

	for _, ch := range chapters {
		flag := ""
		if median > 0 && (float64(ch.words) > 2*median || float64(ch.words) < 0.5*median) {
			flag = "⚠️"
		}
		fmt.Fprintf(&b, "| %s | %d | %s |\n", ch.name, ch.words, flag)
	}

	return b.String()
}

// dialogueRe matches text between straight double quotes ("...") or curly double quotes (\u201c...\u201d).
var dialogueRe = regexp.MustCompile("[\"\u201c]([^\"\u201d]*)[\"\u201d]")

// runDialogueRatio reports dialogue vs. narration word counts per chapter and
// overall, flagging chapters with extreme dialogue ratios (<10% or >70%).
func runDialogueRatio(blobs []RoleText) string {
	type chapterStats struct {
		name           string
		totalWords     int
		dialogueWords  int
		narrationWords int
		dialoguePct    float64
		flagged        bool
	}

	var chapters []chapterStats
	var grandTotalWords, grandDialogueWords int

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		if text == "" {
			continue
		}

		totalWords := len(strings.Fields(text))
		if totalWords == 0 {
			continue
		}

		// Count words inside dialogue matches.
		dialogueWords := 0
		matches := dialogueRe.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			dialogueWords += len(strings.Fields(m[1]))
		}

		narrationWords := totalWords - dialogueWords
		if narrationWords < 0 {
			narrationWords = 0
		}

		pct := 0.0
		if totalWords > 0 {
			pct = float64(dialogueWords) / float64(totalWords) * 100
		}

		flagged := pct < 10 || pct > 70

		chapters = append(chapters, chapterStats{
			name:           strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name)),
			totalWords:     totalWords,
			dialogueWords:  dialogueWords,
			narrationWords: narrationWords,
			dialoguePct:    pct,
			flagged:        flagged,
		})

		grandTotalWords += totalWords
		grandDialogueWords += dialogueWords
	}

	if len(chapters) == 0 {
		return "No manuscript chapters with content found."
	}

	// Manuscript-wide dialogue percentage.
	overallPct := 0.0
	if grandTotalWords > 0 {
		overallPct = float64(grandDialogueWords) / float64(grandTotalWords) * 100
	}
	grandNarrationWords := grandTotalWords - grandDialogueWords

	// Build output.
	var b strings.Builder

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Manuscript dialogue:** %.1f%%\n", overallPct)
	fmt.Fprintf(&b, "- **Total words:** %d\n", grandTotalWords)
	fmt.Fprintf(&b, "- **Dialogue words:** %d\n", grandDialogueWords)
	fmt.Fprintf(&b, "- **Narration words:** %d\n", grandNarrationWords)
	fmt.Fprintf(&b, "- **Chapters analyzed:** %d\n", len(chapters))
	b.WriteString("- **Flag threshold:** <10% or >70% dialogue\n")
	b.WriteString("\n")

	// Per-chapter table.
	b.WriteString("## Per-Chapter Breakdown\n\n")
	b.WriteString("| Chapter | Words | Dialogue | Narration | Dialogue% | Flag |\n")
	b.WriteString("|---|---|---|---|---|---|\n")

	for _, ch := range chapters {
		flag := ""
		if ch.flagged {
			flag = "⚠️"
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %.1f%% | %s |\n",
			ch.name, ch.totalWords, ch.dialogueWords, ch.narrationWords, ch.dialoguePct, flag)
	}

	return b.String()
}

// standardTags is the set of dialogue tags considered standard (invisible to readers).
var standardTags = map[string]bool{
	"said":     true,
	"asked":    true,
	"replied":  true,
	"answered": true,
}

// extendedTags is the set of non-standard dialogue tag verbs (said-bookisms).
var extendedTags = map[string]bool{
	"whispered":   true,
	"shouted":     true,
	"exclaimed":   true,
	"murmured":    true,
	"growled":     true,
	"muttered":    true,
	"cried":       true,
	"screamed":    true,
	"hissed":      true,
	"snapped":     true,
	"sighed":      true,
	"groaned":     true,
	"laughed":     true,
	"stammered":   true,
	"stuttered":   true,
	"pleaded":     true,
	"demanded":    true,
	"insisted":    true,
	"suggested":   true,
	"warned":      true,
	"bellowed":    true,
	"called":      true,
	"declared":    true,
	"announced":   true,
	"remarked":    true,
	"observed":    true,
	"noted":       true,
	"added":       true,
	"continued":   true,
	"admitted":    true,
	"agreed":      true,
	"argued":      true,
	"begged":      true,
	"commanded":   true,
	"complained":  true,
	"conceded":    true,
	"confessed":   true,
	"denied":      true,
	"gasped":      true,
	"grumbled":    true,
	"interrupted": true,
	"lied":        true,
	"mentioned":   true,
	"moaned":      true,
	"objected":    true,
	"offered":     true,
	"ordered":     true,
	"promised":    true,
	"protested":   true,
	"repeated":    true,
	"responded":   true,
	"retorted":    true,
	"snickered":   true,
	"sobbed":      true,
	"urged":       true,
	"whimpered":   true,
	"wondered":    true,
	"yelled":      true,
}

// tagAfterQuoteRe matches a closing quote (straight or curly) followed by optional
// comma/period, whitespace, and up to two words — to detect dialogue tag patterns.
var tagAfterQuoteRe = regexp.MustCompile(`["\x{201d}][,.]?\s+([a-zA-Z]+)(?:\s+([a-zA-Z]+))?`)

// isSpeechVerb checks whether a word is a known speech verb (standard or extended).
func isSpeechVerb(w string) bool {
	lower := strings.ToLower(w)
	return standardTags[lower] || extendedTags[lower]
}

// isStandardTag checks whether a word is one of the standard dialogue tags.
func isStandardTag(w string) bool {
	return standardTags[strings.ToLower(w)]
}

// runDialogueTagAudit scans manuscript text for dialogue tags following quoted
// speech. It tallies standard tags (said, asked, replied, answered) vs.
// non-standard tags and reports counts, percentages, and a frequency table.
// Flags if standard-tag percentage drops below 70%.
func runDialogueTagAudit(blobs []RoleText) string {
	tagCounts := make(map[string]int)
	totalStandard := 0
	totalNonStandard := 0

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		if text == "" {
			continue
		}

		matches := tagAfterQuoteRe.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			word1 := m[1]
			word2 := ""
			if len(m) > 2 {
				word2 = m[2]
			}

			// Check if word1 or word2 is a speech verb.
			// Pattern 1: "..." said / "..." he said
			var foundVerb string
			if isSpeechVerb(word1) {
				foundVerb = strings.ToLower(word1)
			} else if word2 != "" && isSpeechVerb(word2) {
				foundVerb = strings.ToLower(word2)
			}

			if foundVerb == "" {
				continue
			}

			tagCounts[foundVerb]++
			if isStandardTag(foundVerb) {
				totalStandard++
			} else {
				totalNonStandard++
			}
		}
	}

	totalTags := totalStandard + totalNonStandard

	if totalTags == 0 {
		return "No dialogue tags detected in manuscript."
	}

	standardPct := float64(totalStandard) / float64(totalTags) * 100

	// Build output.
	var b strings.Builder

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Total dialogue tags found:** %d\n", totalTags)
	fmt.Fprintf(&b, "- **Standard tags (said, asked, replied, answered):** %d\n", totalStandard)
	fmt.Fprintf(&b, "- **Non-standard tags:** %d\n", totalNonStandard)
	fmt.Fprintf(&b, "- **Standard-tag percentage:** %.1f%%\n", standardPct)

	if standardPct < 70 {
		b.WriteString("- ⚠️ **Standard-tag usage is below 70%** — consider reducing said-bookisms\n")
	}
	b.WriteString("\n")

	// Collect non-standard tags sorted by frequency.
	type tagEntry struct {
		tag   string
		count int
	}

	var nonStandardList []tagEntry
	for tag, count := range tagCounts {
		if !standardTags[tag] {
			nonStandardList = append(nonStandardList, tagEntry{tag: tag, count: count})
		}
	}

	if len(nonStandardList) > 0 {
		sort.Slice(nonStandardList, func(i, j int) bool {
			if nonStandardList[i].count != nonStandardList[j].count {
				return nonStandardList[i].count > nonStandardList[j].count
			}
			return nonStandardList[i].tag < nonStandardList[j].tag
		})

		b.WriteString("## Non-Standard Tags\n\n")
		b.WriteString("| Tag | Count |\n")
		b.WriteString("|---|---|\n")
		for _, entry := range nonStandardList {
			fmt.Fprintf(&b, "| %s | %d |\n", entry.tag, entry.count)
		}
	}

	return b.String()
}

// passiveRe matches a be-verb followed by a regular past participle (ending in -ed or -en).
var passiveRe = regexp.MustCompile(`(?i)\b(was|were|is|are|am|been|being|be)\s+(\w+(ed|en))\b`)

// passiveIrregRe matches a be-verb followed by any single word, used to check against
// the irregular past participles set.
var passiveIrregRe = regexp.MustCompile(`(?i)\b(was|were|is|are|am|been|being|be)\s+([a-zA-Z]+)\b`)

// irregularPastParticiples is the set of common irregular past participles that
// form passive voice when preceded by a be-verb.
var irregularPastParticiples = map[string]bool{
	"done": true, "gone": true, "seen": true, "taken": true, "given": true,
	"written": true, "shown": true, "known": true, "grown": true, "thrown": true,
	"drawn": true, "driven": true, "risen": true, "spoken": true, "chosen": true,
	"frozen": true, "broken": true, "stolen": true, "forgotten": true, "hidden": true,
	"bitten": true, "beaten": true, "shaken": true, "torn": true, "worn": true,
	"born": true, "sworn": true, "begun": true, "sung": true, "rung": true,
	"drunk": true, "sunk": true, "hung": true, "held": true, "kept": true,
	"left": true, "lost": true, "made": true, "met": true, "paid": true,
	"said": true, "sold": true, "sent": true, "shot": true, "shut": true,
	"spent": true, "stood": true, "taught": true, "thought": true, "told": true,
	"understood": true, "built": true, "bought": true, "brought": true,
	"caught": true, "felt": true, "found": true, "got": true, "heard": true,
	"hit": true, "hurt": true, "let": true, "put": true, "read": true,
	"run": true, "set": true, "sat": true, "cut": true, "won": true,
	"woven": true,
}

// isPassiveSentence returns true if the sentence contains at least one passive
// voice construction: a be-verb followed by a past participle (regular -ed/-en
// ending or irregular form).
func isPassiveSentence(sentence string) bool {
	// Check regular past participles via regex.
	if passiveRe.MatchString(sentence) {
		return true
	}
	// Check irregular past participles.
	matches := passiveIrregRe.FindAllStringSubmatch(sentence, -1)
	for _, m := range matches {
		word := strings.ToLower(m[2])
		if irregularPastParticiples[word] {
			return true
		}
	}
	return false
}

// runPassiveVoice reports passive voice density per chapter and manuscript-wide.
// It flags chapters exceeding 10% passive voice sentences.
func runPassiveVoice(blobs []RoleText) string {
	type chapterStats struct {
		name       string
		sentences  int
		passive    int
		passivePct float64
		flagged    bool
	}

	var chapters []chapterStats
	var totalSentences, totalPassive int

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		if text == "" {
			continue
		}

		sents := splitSentences(text)
		if len(sents) == 0 {
			continue
		}

		passiveCount := 0
		for _, s := range sents {
			if isPassiveSentence(s) {
				passiveCount++
			}
		}

		pct := 0.0
		if len(sents) > 0 {
			pct = float64(passiveCount) / float64(len(sents)) * 100
		}

		flagged := pct > 10

		chapters = append(chapters, chapterStats{
			name:       strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name)),
			sentences:  len(sents),
			passive:    passiveCount,
			passivePct: pct,
			flagged:    flagged,
		})

		totalSentences += len(sents)
		totalPassive += passiveCount
	}

	if len(chapters) == 0 {
		return "No manuscript chapters with content found."
	}

	// Manuscript-wide passive percentage.
	overallPct := 0.0
	if totalSentences > 0 {
		overallPct = float64(totalPassive) / float64(totalSentences) * 100
	}

	// Build output.
	var b strings.Builder

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Manuscript passive voice:** %.1f%%\n", overallPct)
	fmt.Fprintf(&b, "- **Total sentences:** %d\n", totalSentences)
	fmt.Fprintf(&b, "- **Passive sentences:** %d\n", totalPassive)
	fmt.Fprintf(&b, "- **Chapters analyzed:** %d\n", len(chapters))
	b.WriteString("- **Flag threshold:** >10% passive voice\n")
	b.WriteString("\n")

	// Per-chapter table.
	b.WriteString("## Per-Chapter Breakdown\n\n")
	b.WriteString("| Chapter | Sentences | Passive | Passive% | Flag |\n")
	b.WriteString("|---|---|---|---|---|\n")

	for _, ch := range chapters {
		flag := ""
		if ch.flagged {
			flag = "⚠️"
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %.1f%% | %s |\n",
			ch.name, ch.sentences, ch.passive, ch.passivePct, flag)
	}

	return b.String()
}

type fictionLine struct {
	Text  string
	Start int // byte offset in original (inclusive)
	End   int // byte offset exclusive
}

// wrapFictionLines word-wraps text to width runes per line, tracking byte offsets.
func wrapFictionLines(text string, width int) []fictionLine {
	if width < 1 {
		width = fictionCharsPerLine
	}
	var lines []fictionLine
	flush := func(start, end int) {
		if start >= end || start < 0 || end > len(text) {
			return
		}
		lines = append(lines, fictionLine{Text: text[start:end], Start: start, End: end})
	}

	i := 0
	for i < len(text) {
		if text[i] == '\n' {
			// Preserve blank lines as empty numbered lines for readability.
			lines = append(lines, fictionLine{Text: "", Start: i, End: i})
			i++
			continue
		}
		lineStart := i
		lineWidth := 0
		lastBreak := -1 // byte index of whitespace where we can wrap
		for i < len(text) && text[i] != '\n' {
			r, size := utf8.DecodeRuneInString(text[i:])
			if r == utf8.RuneError && size == 1 {
				size = 1
			}
			if lineWidth >= width && lineWidth > 0 {
				if lastBreak > lineStart {
					flush(lineStart, lastBreak)
					i = lastBreak
					for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
						i++
					}
					lineStart = i
					lineWidth = 0
					lastBreak = -1
					continue
				}
				// Hard-break overlong tokens; always advance at least one rune.
				if i == lineStart {
					i += size
				}
				flush(lineStart, i)
				lineStart = i
				lineWidth = 0
				lastBreak = -1
				continue
			}
			if r == ' ' || r == '\t' {
				lastBreak = i
			}
			i += size
			lineWidth++
		}
		flush(lineStart, i)
		if i < len(text) && text[i] == '\n' {
			i++
		}
	}
	return lines
}

func lineRangeForOffsets(lines []fictionLine, start, end int) (lineStart, lineEnd int) {
	for i, ln := range lines {
		if ln.End <= start || ln.Start >= end {
			continue
		}
		n := i + 1
		if lineStart == 0 {
			lineStart = n
		}
		lineEnd = n
	}
	if lineStart == 0 {
		for i, ln := range lines {
			if start >= ln.Start && (start < ln.End || ln.Start == ln.End) {
				return i + 1, i + 1
			}
		}
		if len(lines) == 0 {
			return 1, 1
		}
		return 1, 1
	}
	return lineStart, lineEnd
}

// StickyLineSpan is a 1-based inclusive line range covering a sticky sentence.
type StickyLineSpan struct {
	LineStart int `json:"line_start"`
	LineEnd   int `json:"line_end"`
}

// StickyChapterData is one chapter row for sticky_sentences data_json.
type StickyChapterData struct {
	Chapter   string           `json:"chapter"`
	Rel       string           `json:"rel"`
	Sentences int              `json:"sentences"`
	Sticky    int              `json:"sticky"`
	StickyPct float64          `json:"sticky_pct"`
	Spans     []StickyLineSpan `json:"spans"`
}

// StickySentencesData is the structured payload for sticky_sentences.
type StickySentencesData struct {
	Kind      string              `json:"kind"`
	LineWidth int                 `json:"line_width"`
	Chapters  []StickyChapterData `json:"chapters"`
}

// StickyLineView is one numbered line in the chapter sticky dialog.
type StickyLineView struct {
	Line      int    `json:"line"`
	Text      string `json:"text"`
	Highlight bool   `json:"highlight"`
}

// StickyChapterContext is the full chapter view for the sticky dialog.
type StickyChapterContext struct {
	Chapter     string          `json:"chapter"`
	Rel         string          `json:"rel"`
	LineWidth   int             `json:"line_width"`
	StickyCount int             `json:"sticky_count"`
	Lines       []StickyLineView `json:"lines"`
}

func sentenceGlueRatio(s string) (ratio float64, ok bool) {
	words := strings.Fields(s)
	if len(words) == 0 {
		return 0, false
	}
	glueCount := 0
	for _, w := range words {
		if isGlueWord(normalizeWord(w)) {
			glueCount++
		}
	}
	return float64(glueCount) / float64(len(words)), true
}

func stickySpansForChapter(text string, width int, stickyThreshold float64) (spans []StickyLineSpan, stickyCount, sentCount int, top []struct {
	text  string
	ratio float64
}) {
	if stickyThreshold <= 0 {
		stickyThreshold = 0.45
	}
	lines := wrapFictionLines(text, width)
	sents := splitSentences(text)
	sentCount = len(sents)
	pos := 0
	for _, s := range sents {
		ratio, ok := sentenceGlueRatio(s)
		if !ok {
			continue
		}
		idx := strings.Index(text[pos:], s)
		start, end := pos, pos
		if idx >= 0 {
			start = pos + idx
			end = start + len(s)
			pos = end
		}
		top = append(top, struct {
			text  string
			ratio float64
		}{text: s, ratio: ratio})
		if ratio > stickyThreshold {
			stickyCount++
			ls, le := lineRangeForOffsets(lines, start, end)
			spans = append(spans, StickyLineSpan{LineStart: ls, LineEnd: le})
		}
	}
	return spans, stickyCount, sentCount, top
}

func lineHighlighted(spans []StickyLineSpan, lineNum int) bool {
	for _, sp := range spans {
		if lineNum >= sp.LineStart && lineNum <= sp.LineEnd {
			return true
		}
	}
	return false
}

// BuildStickyChapterContext reloads a manuscript chapter and returns fiction-wrapped
// lines with every sticky sentence highlighted.
func BuildStickyChapterContext(projectPath, chapterRel string) (StickyChapterContext, error) {
	root := ResolveProjectRoot(projectPath)
	if root == "" {
		return StickyChapterContext{}, fmt.Errorf("select a book or series first")
	}
	chapterRel = strings.TrimSpace(chapterRel)
	if chapterRel == "" {
		return StickyChapterContext{}, fmt.Errorf("missing chapter")
	}

	var text, name, rel string
	path := filepath.Join(root, filepath.FromSlash(chapterRel))
	if b, err := os.ReadFile(path); err == nil {
		text = string(b)
		name = filepath.Base(path)
		rel = chapterRel
	} else {
		blobs := CollectNeededText(root, []string{"manuscript"})
		for _, blob := range blobs {
			if blob.Rel == chapterRel || filepath.Base(blob.Rel) == filepath.Base(chapterRel) {
				text = blob.Text
				name = blob.Name
				rel = blob.Rel
				break
			}
		}
	}
	if strings.TrimSpace(text) == "" {
		return StickyChapterContext{}, fmt.Errorf("chapter not found")
	}

	text = strings.TrimSpace(text)
	spans, stickyCount, _, _ := stickySpansForChapter(text, fictionCharsPerLine, 0.45)
	lines := wrapFictionLines(text, fictionCharsPerLine)
	out := make([]StickyLineView, 0, len(lines))
	for i, ln := range lines {
		n := i + 1
		out = append(out, StickyLineView{
			Line:      n,
			Text:      ln.Text,
			Highlight: lineHighlighted(spans, n),
		})
	}
	return StickyChapterContext{
		Chapter:     strings.TrimSuffix(name, filepath.Ext(name)),
		Rel:         rel,
		LineWidth:   fictionCharsPerLine,
		StickyCount: stickyCount,
		Lines:       out,
	}, nil
}

// runStickySentences reports sentences overloaded with function words (glue words).
// A sentence is "sticky" when >45% of its words are glue words. Returns markdown
// and a JSON payload with per-chapter sticky line spans for the View dialog.
func runStickySentences(blobs []RoleText, cfg LocalConfig) (markdown string, dataJSON string) {
	stickyThreshold := float64(cfg.GlueThreshold) / 100
	if stickyThreshold <= 0 {
		stickyThreshold = 0.45
	}
	type stickySentence struct {
		text    string
		chapter string
		ratio   float64
	}

	var chapterData []StickyChapterData
	var topSentences []stickySentence
	var totalSentences, totalSticky int

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		if text == "" {
			continue
		}

		chapterName := strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name))
		spans, stickyCount, sentCount, top := stickySpansForChapter(text, fictionCharsPerLine, stickyThreshold)
		if sentCount == 0 {
			continue
		}

		pct := 0.0
		if sentCount > 0 {
			pct = float64(stickyCount) / float64(sentCount) * 100
		}

		chapterData = append(chapterData, StickyChapterData{
			Chapter:   chapterName,
			Rel:       blob.Rel,
			Sentences: sentCount,
			Sticky:    stickyCount,
			StickyPct: pct,
			Spans:     spans,
		})

		for _, t := range top {
			topSentences = append(topSentences, stickySentence{
				text:    t.text,
				chapter: chapterName,
				ratio:   t.ratio,
			})
		}

		totalSentences += sentCount
		totalSticky += stickyCount
	}

	payload := StickySentencesData{
		Kind:      "sticky_sentences",
		LineWidth: fictionCharsPerLine,
		Chapters:  chapterData,
	}
	if raw, err := json.Marshal(payload); err == nil {
		dataJSON = string(raw)
	} else {
		dataJSON = `{"kind":"sticky_sentences","line_width":60,"chapters":[]}`
	}

	if len(chapterData) == 0 {
		return "No manuscript chapters with content found.", dataJSON
	}

	overallPct := 0.0
	if totalSentences > 0 {
		overallPct = float64(totalSticky) / float64(totalSentences) * 100
	}

	sort.Slice(topSentences, func(i, j int) bool {
		return topSentences[i].ratio > topSentences[j].ratio
	})
	if len(topSentences) > 10 {
		topSentences = topSentences[:10]
	}

	var b strings.Builder
	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Total sentences:** %d\n", totalSentences)
	fmt.Fprintf(&b, "- **Sticky sentences:** %d\n", totalSticky)
	fmt.Fprintf(&b, "- **Sticky percentage:** %.1f%%\n", overallPct)
	fmt.Fprintf(&b, "- **Chapters analyzed:** %d\n", len(chapterData))
	fmt.Fprintf(&b, "- **Sticky threshold:** >%.0f%% glue words\n", stickyThreshold*100)
	fmt.Fprintf(&b, "- **View line length:** %d characters (~5.5×8 fiction)\n", fictionCharsPerLine)
	b.WriteString("\n")

	b.WriteString("## Per-Chapter Breakdown\n\n")
	b.WriteString("| Chapter | Sentences | Sticky | Sticky% | View |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for i, ch := range chapterData {
		fmt.Fprintf(&b, "| %s | %d | %d | %.1f%% | [View](#sticky-chapter-%d) |\n",
			ch.Chapter, ch.Sentences, ch.Sticky, ch.StickyPct, i)
	}
	b.WriteString("\n")

	b.WriteString("## Top 10 Stickiest Sentences\n\n")
	b.WriteString("| Sentence | Glue% | Chapter |\n")
	b.WriteString("|---|---|---|\n")
	for _, ts := range topSentences {
		display := ts.text
		if len(display) > 80 {
			display = display[:80] + "..."
		}
		display = strings.ReplaceAll(display, "|", "\\|")
		fmt.Fprintf(&b, "| %s | %.1f%% | %s |\n",
			display, ts.ratio*100, ts.chapter)
	}

	return b.String(), dataJSON
}

// runRepeatedPhrases finds 2-gram, 3-gram, and 4-gram phrases that appear 5+
// times across the manuscript, excludes all-glue-word phrases, and reports the
// top 50 by frequency.
func runRepeatedPhrases(blobs []RoleText) string {
	type phraseEntry struct {
		phrase string
		count  int
		size   int
	}

	// counts maps "n:phrase" → count.
	counts := make(map[string]int)

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		words := strings.Fields(blob.Text)
		normalized := make([]string, 0, len(words))
		for _, w := range words {
			nw := normalizeWord(w)
			if nw != "" {
				normalized = append(normalized, nw)
			}
		}

		for n := 2; n <= 4; n++ {
			for i := 0; i <= len(normalized)-n; i++ {
				key := fmt.Sprintf("%d:%s", n, strings.Join(normalized[i:i+n], " "))
				counts[key]++
			}
		}
	}

	// Filter: count >= 5, not all glue words.
	var results []phraseEntry
	for key, count := range counts {
		if count < 5 {
			continue
		}
		// Parse size from the key prefix (e.g. "2:the quick" → size=2, phrase="the quick").
		colonIdx := strings.Index(key, ":")
		if colonIdx < 0 {
			continue
		}
		size := 0
		fmt.Sscanf(key[:colonIdx], "%d", &size)
		phrase := key[colonIdx+1:]

		// Check if all words are glue words — skip if so.
		phraseWords := strings.Fields(phrase)
		allGlue := true
		for _, pw := range phraseWords {
			if !isGlueWord(pw) {
				allGlue = false
				break
			}
		}
		if allGlue {
			continue
		}

		results = append(results, phraseEntry{phrase: phrase, count: count, size: size})
	}

	// Sort by count descending, then by phrase alphabetically for stability.
	sort.Slice(results, func(i, j int) bool {
		if results[i].count != results[j].count {
			return results[i].count > results[j].count
		}
		return results[i].phrase < results[j].phrase
	})

	// Cap at 50.
	if len(results) > 50 {
		results = results[:50]
	}

	var b strings.Builder
	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("Unique repeated phrases found: **%d**\n\n", len(results)))

	if len(results) == 0 {
		b.WriteString("No phrases found appearing 5 or more times.\n")
		return b.String()
	}

	b.WriteString("## Repeated Phrases\n\n")
	b.WriteString("| Phrase | Count | Size |\n")
	b.WriteString("|---|---|---|\n")

	for _, entry := range results {
		fmt.Fprintf(&b, "| %s | %d | %d-gram |\n", entry.phrase, entry.count, entry.size)
	}

	return b.String()
}

// runParagraphLength reports paragraph length statistics per chapter and
// manuscript-wide, flags very long paragraphs (>200 words) with their
// position, and flags chapters with no short paragraphs (<30 words).
func runParagraphLength(blobs []RoleText) string {
	type longPara struct {
		chapter  string
		position int // 1-based
		total    int
		words    int
	}

	type chapterStats struct {
		name       string
		paragraphs int
		avgLen     float64
		longCount  int
		noShort    bool
	}

	var chapters []chapterStats
	var flagged []longPara
	var totalParagraphs int
	var totalWords int

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		if text == "" {
			continue
		}

		chapterName := strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name))
		paras := splitParagraphs(text)
		if len(paras) == 0 {
			continue
		}

		sum := 0
		longCount := 0
		hasShort := false

		for i, p := range paras {
			wc := len(strings.Fields(p))
			sum += wc

			if wc > 200 {
				longCount++
				flagged = append(flagged, longPara{
					chapter:  chapterName,
					position: i + 1,
					total:    len(paras),
					words:    wc,
				})
			}

			if wc < 30 {
				hasShort = true
			}
		}

		avg := 0.0
		if len(paras) > 0 {
			avg = float64(sum) / float64(len(paras))
		}

		chapters = append(chapters, chapterStats{
			name:       chapterName,
			paragraphs: len(paras),
			avgLen:     avg,
			longCount:  longCount,
			noShort:    !hasShort,
		})

		totalParagraphs += len(paras)
		totalWords += sum
	}

	if len(chapters) == 0 {
		return "No manuscript chapters with content found."
	}

	// Manuscript-wide average paragraph length.
	overallAvg := 0.0
	if totalParagraphs > 0 {
		overallAvg = float64(totalWords) / float64(totalParagraphs)
	}

	// Build output.
	var b strings.Builder

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Total paragraphs:** %d\n", totalParagraphs)
	fmt.Fprintf(&b, "- **Average paragraph length:** %.1f words\n", overallAvg)
	fmt.Fprintf(&b, "- **Long paragraphs (>200 words):** %d\n", len(flagged))
	fmt.Fprintf(&b, "- **Chapters analyzed:** %d\n", len(chapters))
	b.WriteString("\n")

	// Per-chapter table.
	b.WriteString("## Per-Chapter Breakdown\n\n")
	b.WriteString("| Chapter | Paragraphs | Avg Length | Long (>200w) | No Short? |\n")
	b.WriteString("|---|---|---|---|---|\n")

	for _, ch := range chapters {
		noShortFlag := ""
		if ch.noShort {
			noShortFlag = "⚠️"
		}
		fmt.Fprintf(&b, "| %s | %d | %.1f | %d | %s |\n",
			ch.name, ch.paragraphs, ch.avgLen, ch.longCount, noShortFlag)
	}

	// Long paragraphs section.
	if len(flagged) > 0 {
		b.WriteString("\n")
		b.WriteString("## Long Paragraphs\n\n")
		b.WriteString("| Chapter | Position | Words |\n")
		b.WriteString("|---|---|---|\n")

		for _, lp := range flagged {
			fmt.Fprintf(&b, "| %s | paragraph %d of %d | %d |\n",
				lp.chapter, lp.position, lp.total, lp.words)
		}
	}

	return b.String()
}

// closingCueWords are words that signal a forward-looking hook at the end of a chapter.
var closingCueWords = map[string]bool{
	"would": true, "until": true, "tomorrow": true, "later": true, "unless": true,
}

// classifyOpening returns the opening type for a sentence: "dialogue", "action", or "description".
func classifyOpening(sentence string) string {
	trimmed := strings.TrimSpace(sentence)
	if trimmed == "" {
		return "description"
	}
	// Dialogue: starts with straight or curly opening quote.
	if trimmed[0] == '"' || strings.HasPrefix(trimmed, "\u201c") {
		return "dialogue"
	}
	// Action: contains a past-tense verb (word ending in -ed, at least 4 chars).
	for _, w := range strings.Fields(trimmed) {
		clean := normalizeWord(w)
		if len(clean) >= 4 && strings.HasSuffix(clean, "ed") {
			return "action"
		}
	}
	return "description"
}

// closingHooks checks a sentence for hook indicators and returns a list of matched types.
func closingHooks(sentence string) []string {
	var hooks []string
	trimmed := strings.TrimSpace(sentence)
	if trimmed == "" {
		return hooks
	}

	// Question: contains '?'
	if strings.Contains(trimmed, "?") {
		hooks = append(hooks, "question")
	}
	// Ellipsis: ends with '...' or '…'
	if strings.HasSuffix(trimmed, "...") || strings.HasSuffix(trimmed, "\u2026") {
		hooks = append(hooks, "ellipsis")
	}
	// Cue word: any word in the sentence matches a cue word.
	for _, w := range strings.Fields(trimmed) {
		if closingCueWords[strings.ToLower(strings.TrimFunc(w, func(r rune) bool {
			return !unicode.IsLetter(r)
		}))] {
			hooks = append(hooks, "cue word")
			break
		}
	}
	// Dialogue: contains straight or curly quotes.
	if strings.ContainsAny(trimmed, "\"\u201c\u201d") {
		hooks = append(hooks, "dialogue")
	}

	return hooks
}

// runOpeningClosing analyzes the first and last sentences of each chapter to
// report on opening strength and closing hook presence.
func runOpeningClosing(blobs []RoleText) string {
	type chapterResult struct {
		name         string
		openType     string
		openWords    int
		closeHooks   []string
		closeSummary string
	}

	var results []chapterResult
	openCounts := map[string]int{"dialogue": 0, "action": 0, "description": 0}
	hookCount := 0

	for _, blob := range blobs {
		if blob.Role != "manuscript" {
			continue
		}
		text := strings.TrimSpace(blob.Text)
		if text == "" {
			continue
		}

		chapterName := strings.TrimSuffix(blob.Name, filepath.Ext(blob.Name))
		sents := splitSentences(text)
		if len(sents) == 0 {
			continue
		}

		// First sentence analysis.
		firstSent := sents[0]
		openType := classifyOpening(firstSent)
		openWords := len(strings.Fields(firstSent))

		// Last sentence analysis.
		lastSent := sents[len(sents)-1]
		hooks := closingHooks(lastSent)

		hookSummary := "none"
		if len(hooks) > 0 {
			hookSummary = strings.Join(hooks, ", ")
			hookCount++
		}

		results = append(results, chapterResult{
			name:         chapterName,
			openType:     openType,
			openWords:    openWords,
			closeHooks:   hooks,
			closeSummary: hookSummary,
		})
		openCounts[openType]++
	}

	if len(results) == 0 {
		return "No manuscript chapters with content found."
	}

	// Build output.
	var b strings.Builder

	b.WriteString("## Summary\n\n")
	b.WriteString("**Opening types:**\n\n")
	fmt.Fprintf(&b, "- Dialogue: %d\n", openCounts["dialogue"])
	fmt.Fprintf(&b, "- Action: %d\n", openCounts["action"])
	fmt.Fprintf(&b, "- Description: %d\n", openCounts["description"])
	b.WriteString("\n")
	b.WriteString("**Closing hooks:**\n\n")
	fmt.Fprintf(&b, "- Chapters with a hook indicator: %d / %d\n", hookCount, len(results))
	b.WriteString("\n")

	// Per-chapter table.
	b.WriteString("## Per-Chapter Breakdown\n\n")
	b.WriteString("| Chapter | Opening Type | Opening Words | Closing Hooks |\n")
	b.WriteString("|---|---|---|---|\n")

	for _, r := range results {
		fmt.Fprintf(&b, "| %s | %s | %d | %s |\n",
			r.name, r.openType, r.openWords, r.closeSummary)
	}

	return b.String()
}
