package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapFictionLinesWidth(t *testing.T) {
	text := strings.Repeat("word ", 20) // many short words
	lines := wrapFictionLines(strings.TrimSpace(text), 60)
	if len(lines) < 2 {
		t.Fatalf("expected wrap into multiple lines, got %d: %#v", len(lines), lines)
	}
	for i, ln := range lines {
		n := 0
		for range ln.Text {
			n++
		}
		if n > 60 {
			t.Fatalf("line %d width %d > 60: %q", i+1, n, ln.Text)
		}
	}
}

func TestStickySpansAndContextHighlights(t *testing.T) {
	// First sentence is highly sticky (glue-heavy); second is not.
	sticky := "It is of the to in and a that it for with was on as are at be this have from or."
	normal := "Dragons roared above the mountain peaks."
	text := sticky + " " + normal

	spans, stickyCount, sentCount, _ := stickySpansForChapter(text, fictionCharsPerLine)
	if sentCount < 2 {
		t.Fatalf("sentences %d", sentCount)
	}
	if stickyCount < 1 {
		t.Fatalf("expected sticky sentences, got %d", stickyCount)
	}
	if len(spans) != stickyCount {
		t.Fatalf("spans %d sticky %d", len(spans), stickyCount)
	}

	dir := t.TempDir()
	chapDir := filepath.Join(dir, "01_Manuscript", "01_Chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	rel := filepath.ToSlash(filepath.Join("01_Manuscript", "01_Chapters", "01-Test.md"))
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, err := BuildStickyChapterContext(dir, rel)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.StickyCount != stickyCount {
		t.Fatalf("sticky_count %d want %d", ctx.StickyCount, stickyCount)
	}
	if ctx.LineWidth != fictionCharsPerLine {
		t.Fatalf("line_width %d", ctx.LineWidth)
	}
	highlighted := 0
	for _, ln := range ctx.Lines {
		if ln.Highlight {
			highlighted++
		}
	}
	if highlighted == 0 {
		t.Fatal("expected highlighted lines")
	}
}

func TestRunStickySentencesDataJSON(t *testing.T) {
	sticky := "It is of the to in and a that it for with was on as are at be this have from or."
	md, raw := runStickySentences([]RoleText{{
		Role: "manuscript",
		Rel:  "01_Manuscript/01_Chapters/01.md",
		Name: "01.md",
		Text: sticky + " Dragons roared.",
	}})
	if !strings.Contains(md, "| View |") {
		t.Fatalf("missing View column: %s", md)
	}
	if !strings.Contains(md, "#sticky-chapter-0") {
		t.Fatalf("missing view link: %s", md)
	}
	var data StickySentencesData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}
	if data.Kind != "sticky_sentences" || data.LineWidth != 60 {
		t.Fatalf("%+v", data)
	}
	if len(data.Chapters) != 1 {
		t.Fatalf("chapters %d", len(data.Chapters))
	}
	ch := data.Chapters[0]
	if ch.Sticky < 1 || len(ch.Spans) != ch.Sticky {
		t.Fatalf("chapter %+v", ch)
	}
}
