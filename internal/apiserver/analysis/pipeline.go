package analysis

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"loremetry/internal/apiserver/models"
)

const chapterAnalyzeMaxChars = 12000
const maxPriorChars = 24000

type Runner struct {
	Gateway    models.Gateway
	OnProgress func(step, total int, label string)
}

func (r *Runner) progress(step, total int, label string) {
	if r.OnProgress != nil {
		r.OnProgress(step, total, label)
	}
}

func (r *Runner) Run(ctx context.Context, analysisID string, sources []Source) (string, error) {
	if r.Gateway == nil {
		return "", fmt.Errorf("no model gateway")
	}
	profile := ProfileFor(analysisID)
	switch profile {
	case ProfileChunked:
		return r.runChunked(ctx, analysisID, sources)
	case ProfileTwoStep:
		return r.runTwoStep(ctx, analysisID, sources)
	case ProfileByChapter:
		return r.runByChapter(ctx, analysisID, sources)
	default:
		return r.runSingle(ctx, analysisID, sources)
	}
}

func (r *Runner) runSingle(ctx context.Context, analysisID string, sources []Source) (string, error) {
	r.progress(1, 1, "Writing report…")
	src := truncateSources(sources, 12000)
	sys := SystemPrompt(analysisID)
	user := StepPrompt("final", analysisID, src, nil)
	body, _, err := r.Gateway.Complete(ctx, models.TierStrong, sys, user)
	return body, err
}

func (r *Runner) runTwoStep(ctx context.Context, analysisID string, sources []Source) (string, error) {
	sys := SystemPrompt(analysisID)
	src := truncateSources(sources, 12000)
	total := len(src) + 1
	if total < 2 {
		total = 2
	}
	var summaries []string
	for i, s := range src {
		label := fmt.Sprintf("Summarizing %s…", s.Role)
		if s.Name != "" {
			label = fmt.Sprintf("Summarizing %s (%s)…", s.Role, s.Name)
		}
		r.progress(i+1, total, label)
		chunk := []Source{s}
		user := StepPrompt("summarize_role", analysisID, chunk, nil)
		out, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, user)
		if err != nil {
			return "", err
		}
		summaries = append(summaries, fmt.Sprintf("### %s / %s\n\n%s", s.Role, s.Name, out))
	}
	r.progress(total, total, "Writing report…")
	user := StepPrompt("final", analysisID, nil, summaries)
	body, _, err := r.Gateway.Complete(ctx, models.TierStrong, sys, user)
	return body, err
}

func joinManuscriptText(sources []Source) (string, []Source) {
	var parts []string
	var other []Source
	for _, s := range sources {
		if s.Role == "manuscript" {
			if t := strings.TrimSpace(s.Text); t != "" {
				parts = append(parts, t)
			}
			continue
		}
		other = append(other, s)
	}
	return strings.Join(parts, "\n\n"), other
}

func (r *Runner) runChunked(ctx context.Context, analysisID string, sources []Source) (string, error) {
	sys := SystemPrompt(analysisID)
	manuscript, other := joinManuscriptText(sources)
	other = truncateSources(other, 12000)
	chunks := SplitManuscript(manuscript, 10000)
	total := len(other) + len(chunks) + 2
	if total < 1 {
		total = 1
	}
	step := 0
	var roleSummaries []string
	for _, s := range other {
		step++
		label := fmt.Sprintf("Summarizing %s…", s.Role)
		r.progress(step, total, label)
		user := StepPrompt("summarize_role", analysisID, []Source{s}, nil)
		out, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, user)
		if err != nil {
			return "", err
		}
		roleSummaries = append(roleSummaries, fmt.Sprintf("### %s\n\n%s", s.Role, out))
	}
	var chunkSummaries []string
	for i, ch := range chunks {
		step++
		r.progress(step, total, fmt.Sprintf("Summarizing manuscript section %d of %d…", i+1, len(chunks)))
		user := StepPrompt("summarize_chunk", analysisID, []Source{{Role: "manuscript", Name: fmt.Sprintf("section %d", i+1), Text: ch}}, nil)
		out, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, user)
		if err != nil {
			return "", err
		}
		chunkSummaries = append(chunkSummaries, out)
	}
	step++
	r.progress(step, total, "Merging outline…")
	outlineUser := StepPrompt("merge_outline", analysisID, nil, clipPrior(chunkSummaries, maxPriorChars))
	outline, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, outlineUser)
	if err != nil {
		return "", err
	}
	prior := append(roleSummaries, "### Manuscript outline\n\n"+outline)
	step++
	r.progress(step, total, "Writing report…")
	user := StepPrompt("final", analysisID, nil, clipPrior(prior, maxPriorChars))
	body, _, err := r.Gateway.Complete(ctx, models.TierStrong, sys, user)
	return body, err
}

func chapterDisplayName(s Source, index, total int) string {
	name := strings.TrimSpace(s.Name)
	if name != "" {
		name = strings.TrimSuffix(name, filepath.Ext(name))
		if name != "" {
			return name
		}
	}
	return fmt.Sprintf("Chapter %d", index+1)
}

// manuscriptChapters returns one Source per chapter file, or splits a single blob by headings/size.
func manuscriptChapters(sources []Source) (chapters []Source, other []Source) {
	var ms []Source
	for _, s := range sources {
		if s.Role == "manuscript" {
			if strings.TrimSpace(s.Text) == "" {
				continue
			}
			ms = append(ms, s)
			continue
		}
		other = append(other, s)
	}
	if len(ms) == 0 {
		return nil, other
	}
	if len(ms) == 1 {
		parts := SplitManuscript(ms[0].Text, chapterAnalyzeMaxChars)
		if len(parts) > 1 {
			chapters = make([]Source, len(parts))
			for i, p := range parts {
				chapters[i] = Source{
					Role: "manuscript",
					Rel:  ms[0].Rel,
					Name: fmt.Sprintf("section %d", i+1),
					Text: TruncateExcerpt(p, chapterAnalyzeMaxChars),
				}
			}
			return chapters, other
		}
	}
	chapters = make([]Source, len(ms))
	for i, s := range ms {
		chapters[i] = s
		chapters[i].Text = TruncateExcerpt(s.Text, chapterAnalyzeMaxChars)
	}
	return chapters, other
}

func (r *Runner) runByChapter(ctx context.Context, analysisID string, sources []Source) (string, error) {
	sys := SystemPrompt(analysisID)
	chapters, other := manuscriptChapters(sources)
	other = truncateSources(other, 8000)
	if len(chapters) == 0 {
		r.progress(1, 1, "Writing report…")
		user := StepPrompt("final", analysisID, truncateSources(other, 12000), nil)
		body, _, err := r.Gateway.Complete(ctx, models.TierStrong, sys, user)
		return body, err
	}
	total := len(chapters) + 1
	var findings []string
	for i, ch := range chapters {
		label := fmt.Sprintf("Analyzing chapter %d of %d…", i+1, len(chapters))
		if name := chapterDisplayName(ch, i, len(chapters)); name != "" {
			label = fmt.Sprintf("Analyzing %s (%d of %d)…", name, i+1, len(chapters))
		}
		r.progress(i+1, total, label)
		chapterSources := make([]Source, 0, 1+len(other))
		chapterSources = append(chapterSources, other...)
		chapterSources = append(chapterSources, ch)
		user := StepPrompt("analyze_chapter", analysisID, chapterSources, nil)
		out, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, user)
		if err != nil {
			return "", err
		}
		title := chapterDisplayName(ch, i, len(chapters))
		findings = append(findings, fmt.Sprintf("### %s\n\n%s", title, out))
	}
	r.progress(total, total, "Writing report…")
	user := StepPrompt("final", analysisID, other, clipPrior(findings, maxPriorChars))
	body, _, err := r.Gateway.Complete(ctx, models.TierStrong, sys, user)
	return body, err
}

func truncateSources(sources []Source, maxChars int) []Source {
	out := make([]Source, len(sources))
	for i, s := range sources {
		out[i] = s
		if s.Role == "manuscript" {
			out[i].Text = TruncateExcerpt(s.Text, maxChars)
		} else if len(s.Text) > maxChars {
			out[i].Text = TruncateExcerpt(s.Text, maxChars)
		}
	}
	return out
}

func clipPrior(parts []string, maxChars int) []string {
	if maxChars <= 0 || len(parts) == 0 {
		return parts
	}
	total := 0
	for _, p := range parts {
		total += len(p)
	}
	if total <= maxChars {
		return parts
	}
	out := make([]string, len(parts))
	copy(out, parts)
	for total > maxChars {
		longest := 0
		for i := 1; i < len(out); i++ {
			if len(out[i]) > len(out[longest]) {
				longest = i
			}
		}
		if len(out[longest]) <= 200 {
			break
		}
		keep := len(out[longest]) / 2
		if keep < 200 {
			keep = 200
		}
		total -= len(out[longest]) - keep
		out[longest] = TruncateExcerpt(out[longest], keep)
	}
	return out
}
