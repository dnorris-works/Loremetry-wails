package analysis

import (
	"context"
	"fmt"

	"loremetry/internal/apiserver/models"
)

type Runner struct {
	Gateway models.Gateway
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
	src := truncateSources(sources, 12000)
	sys := SystemPrompt(analysisID)
	user := StepPrompt("final", analysisID, src, nil)
	body, _, err := r.Gateway.Complete(ctx, models.TierStrong, sys, user)
	return body, err
}

func (r *Runner) runTwoStep(ctx context.Context, analysisID string, sources []Source) (string, error) {
	sys := SystemPrompt(analysisID)
	var summaries []string
	for _, s := range sources {
		chunk := []Source{s}
		user := StepPrompt("summarize_role", analysisID, chunk, nil)
		out, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, user)
		if err != nil {
			return "", err
		}
		summaries = append(summaries, fmt.Sprintf("### %s / %s\n\n%s", s.Role, s.Name, out))
	}
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
	var roleSummaries []string
	for _, s := range other {
		user := StepPrompt("summarize_role", analysisID, []Source{s}, nil)
		out, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, user)
		if err != nil {
			return "", err
		}
		roleSummaries = append(roleSummaries, fmt.Sprintf("### %s\n\n%s", s.Role, out))
	}
	chunks := SplitManuscript(manuscript, 10000)
	var chunkSummaries []string
	for i, ch := range chunks {
		user := StepPrompt("summarize_chunk", analysisID, []Source{{Role: "manuscript", Name: fmt.Sprintf("section %d", i+1), Text: ch}}, nil)
		out, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, user)
		if err != nil {
			return "", err
		}
		chunkSummaries = append(chunkSummaries, out)
	}
	outlineUser := StepPrompt("merge_outline", analysisID, nil, chunkSummaries)
	outline, _, err := r.Gateway.Complete(ctx, models.TierFast, sys, outlineUser)
	if err != nil {
		return "", err
	}
	prior := append(roleSummaries, "### Manuscript outline\n\n"+outline)
	user := StepPrompt("final", analysisID, nil, prior)
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
