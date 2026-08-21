package analysis

import (
	"context"
	"fmt"

	"loremetry/internal/apiserver/models"
)

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

func (r *Runner) runChunked(ctx context.Context, analysisID string, sources []Source) (string, error) {
	sys := SystemPrompt(analysisID)
	var manuscript string
	var other []Source
	for _, s := range sources {
		if s.Role == "manuscript" {
			manuscript = s.Text
		} else {
			other = append(other, s)
		}
	}
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
	const maxPriorChars = 24000
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
