package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"loremetry/internal/apiserver/analysis"
	"loremetry/internal/apiserver/auth"
	"loremetry/internal/apiserver/billing"
	"loremetry/internal/apiserver/cache"
	"loremetry/internal/apiserver/db"
	"loremetry/internal/apiserver/ledger"
	"loremetry/internal/apiserver/models"
	"loremetry/internal/apiserver/queue"
)

type server struct {
	auth    *auth.Store
	bill    billing.Provider
	ledger  *ledger.Ledger
	gateway models.Gateway
	jobs    *queue.Service
}

func main() {
	gw := models.NewFromEnv()
	sqlDB, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}
	if sqlDB != nil {
		defer sqlDB.Close()
		log.Print("postgres connected")
	}

	var authStore *auth.Store
	var ledgerStore *ledger.Ledger
	var cacheStore *cache.Store
	if sqlDB != nil {
		authStore = auth.NewPostgres(sqlDB)
		ledgerStore = ledger.NewPostgres(sqlDB)
		cacheStore = cache.NewPostgres(sqlDB)
	} else {
		authStore = auth.New()
		ledgerStore = ledger.New()
		cacheStore = cache.NewMemory()
		if os.Getenv("LOREMETRY_DEV_SEED") != "0" {
			log.Print("in-memory auth (set DATABASE_URL for postgres)")
		}
	}

	var jobSvc *queue.Service
	if sqlDB != nil {
		jobSvc = queue.NewPostgres(sqlDB, ledgerStore, cacheStore, gw)
	} else {
		jobSvc = queue.NewMemory(ledgerStore, cacheStore, gw)
	}

	s := &server{
		auth:    authStore,
		bill:    billing.Fake{Site: env("LOREMETRY_SITE", "https://api.loremetry.com")},
		ledger:  ledgerStore,
		gateway: gw,
		jobs:    jobSvc,
	}
	api := http.NewServeMux()
	api.HandleFunc("GET /health", s.health)
	api.HandleFunc("GET /account", s.account)
	api.HandleFunc("POST /auth/device", s.device)
	api.HandleFunc("POST /billing/checkout", s.checkout)
	api.HandleFunc("POST /webhooks/billing", s.webhook)
	api.HandleFunc("POST /analysis/jobs", s.createJob)
	api.HandleFunc("GET /analysis/jobs/{id}", s.getJob)
	api.HandleFunc("POST /analysis/run", s.run)
	api.HandleFunc("GET /billing/fake-checkout", s.fakeCheckoutPage)

	var site http.Handler
	if dir := websiteDir(); dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			site = websiteHandler(dir)
			log.Printf("serving website from %s", dir)
		}
	}

	addr := env("PORT", "")
	if addr == "" {
		addr = env("LOREMETRY_API_ADDR", "5000")
	}
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	log.Printf("loremetry %s", addr)
	log.Fatal(http.ListenAndServe(addr, cors(routeByHost(api, site, s.health))))
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

func (s *server) parseAnalysisRequest(r *http.Request) (string, []analysis.Source, error) {
	var in struct {
		AnalysisID string            `json:"analysis_id"`
		Sources    []analysis.Source `json:"sources"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return "", nil, err
	}
	in.AnalysisID = strings.TrimSpace(in.AnalysisID)
	if in.AnalysisID == "" {
		return "", nil, errors.New("analysis_id required")
	}
	return in.AnalysisID, in.Sources, nil
}

func (s *server) createJob(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	analysisID, sources, err := s.parseAnalysisRequest(r)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	j, err := s.jobs.Enqueue(u.ID, analysisID, sources, u.MaxConcurrentAI)
	if err != nil {
		s.writeJobError(w, err)
		return
	}
	writeJSON(w, 200, jobResponse(j))
}

func (s *server) getJob(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	j, err := s.jobs.Get(u.ID, id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "job not found"})
		return
	}
	writeJSON(w, 200, jobResponse(j))
}

func (s *server) run(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	analysisID, sources, err := s.parseAnalysisRequest(r)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	j, err := s.jobs.Enqueue(u.ID, analysisID, sources, u.MaxConcurrentAI)
	if err != nil {
		s.writeJobError(w, err)
		return
	}
	if j.Cached || j.Status == queue.StatusDone {
		writeJSON(w, 200, map[string]any{"body": j.Body, "credits": j.Credits, "cached": j.Cached})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()
	j, err = s.jobs.Wait(ctx, u.ID, j.ID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			writeJSON(w, 202, jobResponse(j))
			return
		}
		writeJSON(w, 502, map[string]string{"error": err.Error()})
		return
	}
	if j.Status == queue.StatusFailed {
		writeJSON(w, 502, map[string]string{"error": j.Error})
		return
	}
	writeJSON(w, 200, map[string]any{"body": j.Body, "credits": s.ledger.Balance(u.ID)})
}

func (s *server) writeJobError(w http.ResponseWriter, err error) {
	switch {
	case queue.IsCreditsEmpty(err):
		writeJSON(w, 402, map[string]string{"code": "CREDITS_EMPTY", "message": "This month’s credits are used up."})
	case queue.IsConcurrentLimit(err):
		writeJSON(w, 429, map[string]string{"code": "CONCURRENT_LIMIT", "message": "Another analysis is already running. Wait or upgrade your plan."})
	case errors.Is(err, queue.ErrPlanRequired):
		writeJSON(w, 401, map[string]string{"code": "PLAN_REQUIRED", "message": "This analysis uses AI. Choose a plan / add credits."})
	default:
		writeJSON(w, 500, map[string]string{"error": err.Error()})
	}
}

func jobResponse(j *queue.Job) map[string]any {
	if j == nil {
		return map[string]any{}
	}
	out := map[string]any{
		"job_id":     j.ID,
		"status":     j.Status,
		"step":       j.Step,
		"step_total": j.StepTotal,
		"credits":    j.Credits,
	}
	if j.Body != "" {
		out["body"] = j.Body
	}
	if j.Error != "" {
		out["error"] = j.Error
	}
	if j.Cached {
		out["cached"] = true
	}
	return out
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
