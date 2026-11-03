package localai

import (
	"os"
	"testing"
)

func TestPlatformKey(t *testing.T) {
	k := PlatformKey()
	if k == "" {
		t.Fatal("empty platform")
	}
}

func TestPathsAt(t *testing.T) {
	p := pathsAt("/tmp/localai-test-root", "darwin-amd64")
	if p.Platform != "darwin-amd64" {
		t.Fatalf("platform %q", p.Platform)
	}
	if p.Sidecar == "" || p.Model == "" {
		t.Fatal("empty paths")
	}
}

func TestSidecarCacheArgs(t *testing.T) {
	t.Setenv("LOREMETRY_LLAMA_CACHE_RAM", "")
	args := SidecarCacheArgs()
	if len(args) != 2 || args[0] != "--cache-ram" || args[1] != "4096" {
		t.Fatalf("default args=%v", args)
	}
	t.Setenv("LOREMETRY_LLAMA_CACHE_RAM", "0")
	if got := SidecarCacheArgs(); len(got) != 0 {
		t.Fatalf("disable args=%v", got)
	}
	t.Setenv("LOREMETRY_LLAMA_CACHE_RAM", "-1")
	args = SidecarCacheArgs()
	if len(args) != 2 || args[1] != "-1" {
		t.Fatalf("unlimited args=%v", args)
	}
}

func TestPromptCacheRAMMiBEnv(t *testing.T) {
	os.Unsetenv("LOREMETRY_LLAMA_CACHE_RAM")
	if got := PromptCacheRAMMiB(); got != DefaultPromptCacheRAMMiB {
		t.Fatalf("default=%d want %d", got, DefaultPromptCacheRAMMiB)
	}
}
