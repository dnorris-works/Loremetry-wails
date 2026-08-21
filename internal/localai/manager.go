package localai

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"loremetry/internal/apiserver/models"
)

// Status is exposed to the frontend.
type Status struct {
	Ready    bool   `json:"ready"`
	Starting bool   `json:"starting"`
	Error    string `json:"error"`
	Model    string `json:"model"`
	Platform string `json:"platform"`
	BaseURL  string `json:"base_url"`
	Context  int    `json:"context"`
}

// Manager owns the llama-server sidecar process.
type Manager struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	port     int
	paths    BundlePaths
	ready    bool
	starting bool
	lastErr  string
	nCtx     int
}

func NewManager() *Manager {
	return &Manager{port: DefaultPort, nCtx: ContextSize}
}

func (m *Manager) BaseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/v1", m.port)
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Status{
		Ready:    m.ready,
		Starting: m.starting,
		Error:    m.lastErr,
		Model:    ModelID,
		Platform: PlatformKey(),
		BaseURL:  m.BaseURL(),
		Context:  m.nCtx,
	}
}

func (m *Manager) Ready() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ready
}

// StartAsync starts the sidecar in the background when the bundle is present.
func (m *Manager) StartAsync() {
	if !BundlePresent() {
		m.mu.Lock()
		m.lastErr = "local AI bundle not installed in this build"
		m.mu.Unlock()
		return
	}
	go func() {
		_ = m.Start(context.Background())
	}()
}

// Start always (re)launches the sidecar with the current ContextSize.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.starting {
		m.mu.Unlock()
		return m.waitReady(ctx, 120*time.Second)
	}
	m.starting = true
	m.ready = false
	m.lastErr = ""
	m.mu.Unlock()

	err := m.launchAndWait(ctx)

	m.mu.Lock()
	m.starting = false
	if err != nil {
		m.ready = false
		m.lastErr = err.Error()
	} else {
		m.ready = true
		m.lastErr = ""
		m.nCtx = ContextSize
	}
	m.mu.Unlock()
	return err
}

func (m *Manager) launchAndWait(ctx context.Context) error {
	paths, err := ResolveBundle()
	if err != nil {
		return err
	}

	_ = m.killSidecar()
	freeListenPort(m.port)

	m.mu.Lock()
	m.paths = paths
	args := []string{
		"-m", paths.Model,
		"--host", "127.0.0.1",
		"--port", fmt.Sprintf("%d", m.port),
		"-c", fmt.Sprintf("%d", ContextSize),
		"--jinja",
	}
	cmd := exec.Command(paths.Sidecar, args...)
	cmd.Dir = filepath.Dir(paths.Sidecar)
	setSidecarSysProcAttr(cmd)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	libDir := cmd.Dir
	pathEnv := os.Getenv("PATH")
	switch runtime.GOOS {
	case "windows":
		cmd.Env = append(os.Environ(), "PATH="+libDir+string(os.PathListSeparator)+pathEnv)
	case "darwin":
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+libDir)
	default:
		cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+libDir)
	}
	if err := cmd.Start(); err != nil {
		m.mu.Unlock()
		return fmt.Errorf("start llama-server: %w", err)
	}
	m.cmd = cmd
	m.mu.Unlock()

	go func() {
		_ = cmd.Wait()
		m.mu.Lock()
		if m.cmd == cmd {
			m.ready = false
			if m.lastErr == "" {
				msg := strings.TrimSpace(stderrBuf.String())
				if msg == "" {
					msg = "local AI process exited"
				} else if len(msg) > 400 {
					msg = msg[len(msg)-400:]
				}
				m.lastErr = msg
			}
			m.cmd = nil
		}
		m.mu.Unlock()
	}()

	waitCtx, waitCancel := context.WithTimeout(ctx, 120*time.Second)
	defer waitCancel()
	if err := m.pollHealth(waitCtx); err != nil {
		tail := strings.TrimSpace(stderrBuf.String())
		_ = m.killSidecar()
		if tail != "" {
			if len(tail) > 400 {
				tail = tail[len(tail)-400:]
			}
			return fmt.Errorf("%w (%s)", err, tail)
		}
		return err
	}
	return nil
}

func freeListenPort(port int) {
	// Orphaned llama-server can keep the old -c value on this port.
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
	if err != nil {
		return
	}
	_ = c.Close()
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("cmd", "/C", fmt.Sprintf("for /f \"tokens=5\" %%a in ('netstat -ano ^| findstr :%d') do taskkill /F /PID %%a", port)).Run()
	default:
		out, err := exec.Command("lsof", "-nP", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-t").Output()
		if err != nil {
			return
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			_ = exec.Command("kill", "-9", line).Run()
		}
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err != nil {
			return
		}
		_ = c.Close()
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *Manager) pollHealth(ctx context.Context) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/health", m.port)
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		res, err := client.Do(req)
		if err == nil {
			_ = res.Body.Close()
			if res.StatusCode >= 200 && res.StatusCode < 500 {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("local AI did not become ready: %w", ctx.Err())
		case <-time.After(400 * time.Millisecond):
		}
	}
}

func (m *Manager) waitReady(ctx context.Context, max time.Duration) error {
	deadline := time.Now().Add(max)
	for {
		m.mu.Lock()
		ready := m.ready
		starting := m.starting
		errMsg := m.lastErr
		m.mu.Unlock()
		if ready {
			return nil
		}
		if !starting && errMsg != "" {
			return fmt.Errorf("%s", errMsg)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for local AI")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
}

// WaitReady waits until ready or returns the last start error.
func (m *Manager) WaitReady(ctx context.Context, max time.Duration) error {
	return m.waitReady(ctx, max)
}

// Stop terminates the sidecar.
func (m *Manager) Stop() error {
	m.mu.Lock()
	m.starting = false
	m.mu.Unlock()
	return m.killSidecar()
}

func (m *Manager) killSidecar() error {
	m.mu.Lock()
	cmd := m.cmd
	m.cmd = nil
	m.ready = false
	m.mu.Unlock()
	killProcessTree(cmd)
	freeListenPort(m.port)
	return nil
}

// MaxPromptChars caps each user message so a buggy caller cannot blow past n_ctx.
const MaxPromptChars = 24000

// Gateway returns an OpenAI-compatible client pointed at the sidecar.
func (m *Manager) Gateway() models.Gateway {
	inner := models.NewOpenAICompat(m.BaseURL(), "local", ModelID, ModelID, 0.4)
	return &cappedGateway{inner: inner, maxUserChars: MaxPromptChars}
}

type cappedGateway struct {
	inner        models.Gateway
	maxUserChars int
}

func (g *cappedGateway) Complete(ctx context.Context, tier models.Tier, system, user string) (string, models.Usage, error) {
	if g.maxUserChars > 0 && len(user) > g.maxUserChars {
		user = user[:g.maxUserChars] + "\n\n[… truncated for local model context …]\n"
	}
	return g.inner.Complete(ctx, tier, system, user)
}

// EnsureStarted tries to start if needed and waits for readiness.
func (m *Manager) EnsureStarted(ctx context.Context) error {
	st := m.Status()
	if st.Ready && st.Context == ContextSize {
		return nil
	}
	if st.Starting {
		return m.WaitReady(ctx, 120*time.Second)
	}
	return m.Start(ctx)
}

// BundlePresent reports whether the install includes local AI files.
func BundlePresent() bool {
	_, err := ResolveBundle()
	return err == nil
}
