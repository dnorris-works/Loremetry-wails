package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Gateway interface {
	Complete(ctx context.Context, system, user string) (string, error)
}

type Stub struct{}

func (Stub) Complete(_ context.Context, _, user string) (string, error) {
	snip := strings.TrimSpace(user)
	if len(snip) > 240 {
		snip = snip[:240] + "…"
	}
	return "# AI analysis (stub)\n\nNo model key is configured. This is a placeholder report so the desktop path can be tested.\n\n" + snip + "\n", nil
}

type OpenAICompat struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

func NewOpenAICompat(baseURL, apiKey, model string) *OpenAICompat {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAICompat{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

func (g *OpenAICompat) Complete(ctx context.Context, system, user string) (string, error) {
	payload := map[string]any{
		"model": g.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.BaseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if g.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.APIKey)
	}
	res, err := g.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("model gateway %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty model response")
	}
	return parsed.Choices[0].Message.Content, nil
}
