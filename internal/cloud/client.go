package cloud

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

const (
	AppSettingBaseURL = "cloud_api_base_url"
	SettingToken      = "cloud_api_token"
	DefaultBaseURL    = "https://api.loremetry.com"
)

type Account struct {
	Plan      string `json:"plan"`
	Credits   int    `json:"credits"`
	Remaining string `json:"remaining"`
}

type CheckoutResponse struct {
	URL string `json:"url"`
}

type RoleText struct {
	Role string `json:"role"`
	Rel  string `json:"rel"`
	Name string `json:"name"`
	Text string `json:"text"`
}

type RunRequest struct {
	AnalysisID string     `json:"analysis_id"`
	Sources    []RoleText `json:"sources"`
}

type RunResponse struct {
	Body    string `json:"body"`
	Credits int    `json:"credits"`
	Cached  bool   `json:"cached"`
}

type JobStatus struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Step      int    `json:"step"`
	StepTotal int    `json:"step_total"`
	Message   string `json:"message,omitempty"`
	Body      string `json:"body"`
	Error     string `json:"error"`
	Credits   int    `json:"credits"`
	Cached    bool   `json:"cached"`
}

type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e APIError) Error() string {
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	return e.Message
}

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func New(baseURL, token string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		BaseURL: baseURL,
		Token:   strings.TrimSpace(token),
		HTTP:    &http.Client{Timeout: 600 * time.Second},
	}
}

func (c *Client) GetAccount() (Account, error) {
	var acc Account
	err := c.do("GET", "/account", nil, &acc)
	return acc, err
}

func RegisterDevice(baseURL, email string) (string, error) {
	cli := New(baseURL, "")
	var out struct {
		Token string `json:"token"`
	}
	if err := cli.postPublic("/auth/device", map[string]string{"email": email}, &out); err != nil {
		return "", err
	}
	token := strings.TrimSpace(out.Token)
	if token == "" {
		return "", fmt.Errorf("no device token returned")
	}
	return token, nil
}

func (c *Client) Checkout() (CheckoutResponse, error) {
	var out CheckoutResponse
	err := c.do("POST", "/billing/checkout", map[string]string{"plan": "writer"}, &out)
	return out, err
}

func (c *Client) SubmitAnalysisJob(analysisID string, sources []RoleText) (JobStatus, error) {
	var out JobStatus
	err := c.do("POST", "/analysis/jobs", RunRequest{AnalysisID: analysisID, Sources: sources}, &out)
	return out, err
}

func (c *Client) GetAnalysisJob(jobID string) (JobStatus, error) {
	var out JobStatus
	err := c.do("GET", "/analysis/jobs/"+jobID, nil, &out)
	return out, err
}

func (c *Client) RunAnalysis(analysisID string, sources []RoleText) (RunResponse, error) {
	job, err := c.SubmitAnalysisJob(analysisID, sources)
	if err != nil {
		return RunResponse{}, err
	}
	if job.Cached || job.Status == "done" {
		return RunResponse{Body: job.Body, Credits: job.Credits, Cached: job.Cached}, nil
	}
	job, err = c.waitJob(context.Background(), job.JobID)
	if err != nil {
		return RunResponse{}, err
	}
	if job.Status == "failed" {
		if job.Error != "" {
			return RunResponse{}, fmt.Errorf("%s", job.Error)
		}
		return RunResponse{}, fmt.Errorf("analysis failed")
	}
	return RunResponse{Body: job.Body, Credits: job.Credits, Cached: job.Cached}, nil
}

func (c *Client) waitJob(ctx context.Context, jobID string) (JobStatus, error) {
	for {
		job, err := c.GetAnalysisJob(jobID)
		if err != nil {
			return JobStatus{}, err
		}
		if job.Status == "done" || job.Status == "failed" {
			return job, nil
		}
		select {
		case <-ctx.Done():
			return job, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *Client) do(method, path string, body any, dest any) error {
	if c.Token == "" {
		return APIError{Status: 401, Code: "PLAN_REQUIRED", Message: "This analysis uses AI. Choose a plan / add credits."}
	}
	return c.request(method, path, body, dest, true)
}

func (c *Client) postPublic(path string, body any, dest any) error {
	return c.request(http.MethodPost, path, body, dest, false)
}

func (c *Client) request(method, path string, body any, dest any, auth bool) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, rdr)
	if err != nil {
		return err
	}
	if auth {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("cloud API: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode == 402 {
		msg := parseMessage(raw, "This month’s credits are used up.")
		code := parseCode(raw, "CREDITS_EMPTY")
		return APIError{Status: 402, Code: code, Message: msg}
	}
	if res.StatusCode == 401 {
		return APIError{Status: 401, Code: "PLAN_REQUIRED", Message: parseMessage(raw, "This analysis uses AI. Choose a plan / add credits.")}
	}
	if res.StatusCode == 429 {
		code := parseCode(raw, "CONCURRENT_LIMIT")
		return APIError{Status: 429, Code: code, Message: parseMessage(raw, "Another analysis is already running.")}
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%s", parseMessage(raw, fmt.Sprintf("cloud API %d", res.StatusCode)))
	}
	if dest == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dest)
}

func parseMessage(raw []byte, fallback string) string {
	var m struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    string `json:"code"`
	}
	if json.Unmarshal(raw, &m) != nil {
		s := strings.TrimSpace(string(raw))
		if s == "" {
			return fallback
		}
		return s
	}
	if m.Message != "" {
		return m.Message
	}
	if m.Error != "" {
		return m.Error
	}
	return fallback
}

func parseCode(raw []byte, fallback string) string {
	var m struct {
		Code string `json:"code"`
	}
	if json.Unmarshal(raw, &m) != nil || m.Code == "" {
		return fallback
	}
	return m.Code
}
