package models

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAICompatCachedPromptTokens(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "ok"}},
			},
			"usage": map[string]any{
				"prompt_tokens":     1200,
				"completion_tokens": 50,
				"total_tokens":      1250,
				"prompt_tokens_details": map[string]any{
					"cached_tokens": 1024,
				},
			},
		})
	}))
	defer srv.Close()

	gw := NewOpenAICompat(srv.URL, "key", "fast", "strong", 0.4)
	body, usage, err := gw.Complete(context.Background(), TierFast, "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if body != "ok" {
		t.Fatalf("body=%q", body)
	}
	if usage.CachedPromptTokens != 1024 {
		t.Fatalf("cached=%d want 1024", usage.CachedPromptTokens)
	}
	if usage.PromptTokens != 1200 {
		t.Fatalf("prompt=%d", usage.PromptTokens)
	}
}
