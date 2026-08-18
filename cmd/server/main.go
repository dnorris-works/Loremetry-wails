package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"loremetry/internal/apiserver/analysis"
	"loremetry/internal/apiserver/auth"
	"loremetry/internal/apiserver/billing"
	"loremetry/internal/apiserver/ledger"
	"loremetry/internal/apiserver/models"
)

type server struct {
	auth    *auth.Store
	bill    billing.Provider
	ledger  *ledger.Ledger
	gateway models.Gateway
}

func main() {
	gw := models.Gateway(models.Stub{})
	base := strings.TrimSpace(os.Getenv("LOREMETRY_MODEL_BASE_URL"))
	key := strings.TrimSpace(os.Getenv("LOREMETRY_MODEL_API_KEY"))
	if base != "" {
		gw = models.NewOpenAICompat(base, key, os.Getenv("LOREMETRY_MODEL"))
	}
	s := &server{
		auth:    auth.New(),
		bill:    billing.Fake{Site: env("LOREMETRY_SITE", "https://loremetry.com")},
		ledger:  ledger.New(),
		gateway: gw,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/v1/account", s.account)
	mux.HandleFunc("POST /api/v1/auth/device", s.device)
	mux.HandleFunc("POST /api/v1/billing/checkout", s.checkout)
	mux.HandleFunc("POST /api/v1/webhooks/billing", s.webhook)
	mux.HandleFunc("POST /api/v1/analysis/run", s.run)
	mux.HandleFunc("GET /api/billing/fake-checkout", s.fakeCheckoutPage)
	addr := env("PORT", "")
	if addr == "" {
		addr = env("LOREMETRY_API_ADDR", "8080")
	}
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	log.Printf("loremetry api %s", addr)
	log.Fatal(http.ListenAndServe(addr, cors(mux)))
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *server) user(w http.ResponseWriter, r *http.Request) (auth.User, bool) {
	u, ok := s.auth.Lookup(auth.Bearer(r))
	if !ok {
		writeJSON(w, 401, map[string]string{"code": "PLAN_REQUIRED", "message": "This analysis uses AI. Choose a plan / add credits."})
		return auth.User{}, false
	}
	return u, true
}

func (s *server) account(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	credits := s.ledger.Balance(u.ID)
	writeJSON(w, 200, map[string]any{
		"plan":      u.Plan,
		"credits":   credits,
		"remaining": creditsLabel(credits),
	})
}

func creditsLabel(n int) string {
	if n <= 0 {
		return "0 credits"
	}
	if n == 1 {
		return "1 credit"
	}
	return strconv.Itoa(n) + " credits"
}

func (s *server) device(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	token, u := s.auth.Issue(in.Email)
	s.ledger.Add(u.ID, 100)
	writeJSON(w, 200, map[string]any{"token": token, "user_id": u.ID, "email": u.Email})
}

func (s *server) checkout(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	var in struct {
		Plan string `json:"plan"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	url, err := s.bill.CheckoutURL(u.ID, in.Plan)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"url": url})
}

func (s *server) webhook(w http.ResponseWriter, r *http.Request) {
	raw, _ := json.Marshal(map[string]string{"ok": "ignored"})
	grant, err := s.bill.ApplyWebhook(raw)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if grant.UserID != "" && grant.Credits > 0 {
		s.ledger.Add(grant.UserID, grant.Credits)
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *server) run(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	var in struct {
		AnalysisID string            `json:"analysis_id"`
		Sources    []analysis.Source `json:"sources"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	in.AnalysisID = strings.TrimSpace(in.AnalysisID)
	if in.AnalysisID == "" {
		writeJSON(w, 400, map[string]string{"error": "analysis_id required"})
		return
	}
	left, ok := s.ledger.Debit(u.ID, in.AnalysisID)
	if !ok {
		writeJSON(w, 402, map[string]string{"code": "CREDITS_EMPTY", "message": "This month’s credits are used up."})
		return
	}
	body, err := s.gateway.Complete(r.Context(), analysis.SystemPrompt(in.AnalysisID), analysis.UserPrompt(in.AnalysisID, in.Sources))
	if err != nil {
		s.ledger.Add(u.ID, s.ledger.CreditCost(in.AnalysisID))
		writeJSON(w, 502, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"body": body, "credits": left})
}

func (s *server) fakeCheckoutPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><body><p>Fake checkout (billing vendor not chosen). Use device token <code>dev-token</code> in the desktop app.</p></body></html>`))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
