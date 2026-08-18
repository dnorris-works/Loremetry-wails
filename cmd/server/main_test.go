package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"loremetry/internal/apiserver/auth"
	"loremetry/internal/apiserver/billing"
	"loremetry/internal/apiserver/ledger"
	"loremetry/internal/apiserver/models"
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

func TestIsAPIHost(t *testing.T) {
	if !isAPIHost("api.loremetry.com") || !isAPIHost("api.loremetry.com:443") {
		t.Fatal("api host")
	}
	if isAPIHost("loremetry.com") || isAPIHost("www.loremetry.com") {
		t.Fatal("site host")
	}
}

func TestRouteByHost(t *testing.T) {
	s := testServer()
	api := http.NewServeMux()
	api.HandleFunc("GET /account", s.account)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>site</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := routeByHost(api, websiteHandler(dir), s.health)

	siteReq := httptest.NewRequest(http.MethodGet, "/", nil)
	siteReq.Host = "loremetry.com"
	siteRec := httptest.NewRecorder()
	h.ServeHTTP(siteRec, siteReq)
	if !strings.Contains(siteRec.Body.String(), "site") {
		t.Fatalf("site: %s", siteRec.Body.String())
	}

	apiReq := httptest.NewRequest(http.MethodGet, "/account", nil)
	apiReq.Host = "api.loremetry.com"
	apiRec := httptest.NewRecorder()
	h.ServeHTTP(apiRec, apiReq)
	if apiRec.Code != 401 {
		t.Fatalf("api status %d body %s", apiRec.Code, apiRec.Body.String())
	}

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthReq.Host = "loremetry.com"
	healthRec := httptest.NewRecorder()
	h.ServeHTTP(healthRec, healthReq)
	if !strings.Contains(healthRec.Body.String(), `"status":"ok"`) {
		t.Fatalf("health: %s", healthRec.Body.String())
	}
}
