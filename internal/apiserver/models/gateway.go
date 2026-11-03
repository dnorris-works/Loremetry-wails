package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Tier int

const (
	TierFast Tier = iota
	TierStrong
)

type Usage struct {
	PromptTokens       int
	CompletionTokens   int
	TotalTokens        int
	CachedPromptTokens int
}

type Gateway interface {
	Complete(ctx context.Context, tier Tier, system, user string) (string, Usage, error)
}

type Stub struct{}

func (Stub) Complete(_ context.Context, _ Tier, _, user string) (string, Usage, error) {
	snip := strings.TrimSpace(user)
	if len(snip) > 240 {
		snip = snip[:240] + "…"
	}
	return "# AI analysis (stub)\n\nNo model key is configured. This is a placeholder report so the desktop path can be tested.\n\n" + snip + "\n", Usage{}, nil
}

type OpenAICompat struct {
	BaseURL    string
	APIKey     string
	FastModel  string
	StrongModel string
	Temperature float64
	HTTP       *http.Client
}

const (
	defaultFireworksBase = "https://api.fireworks.ai/inference/v1"
	defaultFastModel     = "accounts/fireworks/models/gpt-oss-20b"
	defaultStrongModel   = "accounts/fireworks/models/gpt-oss-120b"
)

func NewFromEnv() Gateway {
	key := strings.TrimSpace(os.Getenv("LOREMETRY_MODEL_API_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("FIREWORKS_API_KEY"))
	}
	base := strings.TrimSpace(os.Getenv("LOREMETRY_MODEL_BASE_URL"))
	if base == "" && key != "" {
		base = defaultFireworksBase
	}
	if base == "" {
		return Stub{}
	}
	fast := strings.TrimSpace(os.Getenv("LOREMETRY_MODEL_FAST"))
	if fast == "" {
		fast = strings.TrimSpace(os.Getenv("LOREMETRY_MODEL"))
	}
	if fast == "" {
		fast = defaultFastModel
	}
	strong := strings.TrimSpace(os.Getenv("LOREMETRY_MODEL_STRONG"))
	if strong == "" {
		strong = defaultStrongModel
	}
	return NewOpenAICompat(base, key, fast, strong, 0.4)
}

func NewOpenAICompat(baseURL, apiKey, fastModel, strongModel string, temperature float64) *OpenAICompat {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if temperature <= 0 {
		temperature = 0.4
	}
	if fastModel == "" {
		fastModel = defaultFastModel
	}
	if strongModel == "" {
		strongModel = defaultStrongModel
	}
	return &OpenAICompat{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		FastModel:   fastModel,
		StrongModel: strongModel,
		Temperature: temperature,
		HTTP:        &http.Client{Timeout: 300 * time.Second},
	}
}

func (g *OpenAICompat) model(t Tier) string {
	if t == TierFast {
		return g.FastModel
	}
	return g.StrongModel
}

func (g *OpenAICompat) Complete(ctx context.Context, tier Tier, system, user string) (string, Usage, error) {
	var lastErr error
	backoff := 500 * time.Millisecond
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", Usage{}, ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
		body, usage, err := g.completeOnce(ctx, tier, system, user)
		if err == nil {
			return body, usage, nil
		}
		lastErr = err
		if !retryable(err) {
			break
		}
	}
	return "", Usage{}, lastErr
}

func retryable(err error) bool {
	s := err.Error()
	return strings.Contains(s, "429") || strings.Contains(s, "503")
}

func (g *OpenAICompat) completeOnce(ctx context.Context, tier Tier, system, user string) (string, Usage, error) {
	payload := map[string]any{
		"model":       g.model(tier),
		"temperature": g.Temperature,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", Usage{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.BaseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if g.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.APIKey)
	}
	res, err := g.HTTP.Do(req)
	if err != nil {
		return "", Usage{}, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", Usage{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", Usage{}, fmt.Errorf("model gateway %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
			PromptTokensDetails *struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"prompt_tokens_details"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", Usage{}, err
	}
	if len(parsed.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("empty model response")
	}
	cached := 0
	if parsed.Usage.PromptTokensDetails != nil {
		cached = parsed.Usage.PromptTokensDetails.CachedTokens
	}
	usage := Usage{
		PromptTokens:       parsed.Usage.PromptTokens,
		CompletionTokens:   parsed.Usage.CompletionTokens,
		TotalTokens:        parsed.Usage.TotalTokens,
		CachedPromptTokens: cached,
	}
	return parsed.Choices[0].Message.Content, usage, nil
}
