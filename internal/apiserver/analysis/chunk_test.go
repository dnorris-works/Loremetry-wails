package analysis

import "testing"

func TestSplitManuscriptBySize(t *testing.T) {
	text := string(make([]byte, 25000))
	for i := range text {
		text = text[:i] + "a" + text[i+1:]
	}
	chunks := SplitManuscript(text, 10000)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
}

func TestProfileForChunked(t *testing.T) {
	if ProfileFor("analysis") != ProfileChunked {
		t.Fatal("expected chunked profile for analysis")
	}
	if ProfileFor("genre_analysis") != ProfileSingle {
		t.Fatal("expected single profile for genre_analysis")
	}
}
