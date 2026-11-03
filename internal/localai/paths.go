package localai

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	// ModelFileName is the GGUF shipped inside the app install (no runtime download).
	ModelFileName = "Qwen2.5-3B-Instruct-Q4_K_M.gguf"
	// ModelID is the OpenAI-compat model string sent to llama-server.
	ModelID = "qwen2.5-3b-instruct"
	DefaultPort = 18765
	ContextSize = 32768
	// DefaultPromptCacheRAMMiB caps host RAM used for llama-server KV prompt cache.
	DefaultPromptCacheRAMMiB = 4096
)

// PromptCacheRAMMiB returns MiB for --cache-ram. LOREMETRY_LLAMA_CACHE_RAM overrides;
// use -1 for no limit, 0 to omit the flag (llama-server default).
func PromptCacheRAMMiB() int {
	raw := strings.TrimSpace(os.Getenv("LOREMETRY_LLAMA_CACHE_RAM"))
	if raw == "" {
		return DefaultPromptCacheRAMMiB
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return DefaultPromptCacheRAMMiB
	}
	return n
}

// SidecarCacheArgs returns llama-server flags for host RAM KV prompt caching.
func SidecarCacheArgs() []string {
	ram := PromptCacheRAMMiB()
	if ram == 0 {
		return nil
	}
	return []string{"--cache-ram", strconv.Itoa(ram)}
}

// PlatformKey returns e.g. darwin-arm64, darwin-amd64, windows-amd64.
func PlatformKey() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

func sidecarBinaryName() string {
	switch runtime.GOOS {
	case "windows":
		return "llama-server.exe"
	default:
		return "llama-server"
	}
}

// BundlePaths holds resolved on-disk paths for the sidecar and GGUF.
type BundlePaths struct {
	Root     string
	Sidecar  string
	Model    string
	Platform string
}

func (p BundlePaths) OK() error {
	if p.Sidecar == "" || p.Model == "" {
		return fmt.Errorf("local AI bundle not found for %s", PlatformKey())
	}
	if _, err := os.Stat(p.Sidecar); err != nil {
		return fmt.Errorf("local AI sidecar missing: %s", p.Sidecar)
	}
	if _, err := os.Stat(p.Model); err != nil {
		return fmt.Errorf("local AI model missing: %s", p.Model)
	}
	return nil
}

// ResolveBundle finds the platform sidecar and shared GGUF.
// Search order: LOREMETRY_LOCALAI_DIR, app Resources/localai, exe-adjacent localai, repo third_party (dev).
func ResolveBundle() (BundlePaths, error) {
	plat := PlatformKey()
	candidates := bundleRoots()
	var last error
	for _, root := range candidates {
		p := pathsAt(root, plat)
		if err := p.OK(); err != nil {
			last = err
			continue
		}
		return p, nil
	}
	if last != nil {
		return BundlePaths{Platform: plat}, last
	}
	return BundlePaths{Platform: plat}, fmt.Errorf("local AI bundle not found for %s", plat)
}

func pathsAt(root, plat string) BundlePaths {
	bin := filepath.Join(root, "llama-server", plat, sidecarBinaryName())
	model := filepath.Join(root, "models", ModelFileName)
	// Also allow flat layout: root/llama-server + root/models/...
	if _, err := os.Stat(bin); err != nil {
		alt := filepath.Join(root, sidecarBinaryName())
		if _, err2 := os.Stat(alt); err2 == nil {
			bin = alt
		}
	}
	return BundlePaths{Root: root, Sidecar: bin, Model: model, Platform: plat}
}

func bundleRoots() []string {
	var roots []string
	if env := os.Getenv("LOREMETRY_LOCALAI_DIR"); env != "" {
		roots = append(roots, env)
	}
	if exe, err := os.Executable(); err == nil {
		exe, _ = filepath.EvalSymlinks(exe)
		dir := filepath.Dir(exe)
		// macOS .app: Contents/MacOS/loremetry → Contents/Resources/localai
		if filepath.Base(dir) == "MacOS" {
			resources := filepath.Join(filepath.Dir(dir), "Resources", "localai")
			roots = append(roots, resources)
		}
		roots = append(roots, filepath.Join(dir, "localai"))
		roots = append(roots, filepath.Join(dir, "third_party"))
	}
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, filepath.Join(cwd, "third_party"))
		// wails dev often runs from project root or build/bin
		roots = append(roots, filepath.Join(cwd, "..", "third_party"))
		roots = append(roots, filepath.Join(cwd, "..", "..", "third_party"))
	}
	return roots
}
