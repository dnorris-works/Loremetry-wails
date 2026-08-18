package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"loremetry/server/internal/auth"
	"loremetry/server/internal/billing"
	"loremetry/server/internal/ledger"
	"loremetry/server/internal/models"
)

func testServer() *server {
	return &server{
		auth:    auth.New(),
		bill:    billing.Fake{Site: "http://example"},
		ledger:  ledger.New(),
		gateway: models.Stub{},
	}
}

func TestRunRequiresCredits(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodPost, "/v1/analysis/run", strings.NewReader(`{"analysis_id":"genre_analysis"}`))
	req.Header.Set("Authorization", "Bearer empty-token")
	rec := httptest.NewRecorder()
	s.run(rec, req)
	if rec.Code != 402 {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestRunStubOK(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodPost, "/v1/analysis/run", strings.NewReader(`{"analysis_id":"genre_analysis","sources":[{"role":"manuscript","name":"a.md","text":"hello"}]}`))
	req.Header.Set("Authorization", "Bearer dev-token")
	rec := httptest.NewRecorder()
	s.run(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || !strings.Contains(out.Body, "stub") {
		t.Fatalf("%s %v", rec.Body.String(), err)
	}
}
