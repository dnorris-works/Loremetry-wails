package localai

import "testing"

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
