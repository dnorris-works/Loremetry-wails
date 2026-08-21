package analysis

import (
	"strings"
	"testing"
)

func TestChapterNoteSpecPOV(t *testing.T) {
	s := ChapterNoteSpecFor("pov_discipline")
	want := []string{"POV", "Head-hops", "Info leaks", "Severity"}
	if len(s.RequiredHeadings) != len(want) {
		t.Fatalf("headings=%v", s.RequiredHeadings)
	}
	for i, h := range want {
		if s.RequiredHeadings[i] != h {
			t.Fatalf("heading %d = %q want %q", i, s.RequiredHeadings[i], h)
		}
	}
	if !strings.Contains(s.ChapterInstruction, "POV") {
		t.Fatal(s.ChapterInstruction)
	}
	if !strings.Contains(s.FinalInstruction, "Merge the chapter notes") {
		t.Fatal(s.FinalInstruction)
	}
}

func TestStepPromptAnalyzeChapterIncludesHeadings(t *testing.T) {
	user := StepPrompt("analyze_chapter", "pov_discipline", []Source{{Role: "manuscript", Name: "01.md", Text: "Hello"}}, nil)
	for _, h := range []string{"## POV", "## Head-hops", "## Info leaks", "## Severity"} {
		if !strings.Contains(user, h) {
			t.Fatalf("missing %q in:\n%s", h, user)
		}
	}
	if !strings.Contains(user, "Required headings") {
		t.Fatal(user)
	}
	if !strings.Contains(user, "Hello") {
		t.Fatal("chapter text missing")
	}
}

func TestStepPromptFinalUsesSpec(t *testing.T) {
	user := StepPrompt("final", "show_dont_tell", nil, []string{"### Ch1\n\nnotes"})
	if !strings.Contains(user, "Show Don't Tell") && !strings.Contains(user, "Show Don") {
		// Instruction says "Produce a Show Don't Tell report."
		if !strings.Contains(user, "Produce a Show") {
			t.Fatal(user)
		}
	}
	if !strings.Contains(user, "do not invent") && !strings.Contains(user, "Do not invent") {
		t.Fatal(user)
	}
	if !strings.Contains(user, "### Ch1") {
		t.Fatal(user)
	}
}

func TestChapterNoteSpecMarketing(t *testing.T) {
	s := ChapterNoteSpecFor("genre_analysis")
	if len(s.RequiredHeadings) < 3 {
		t.Fatalf("%v", s.RequiredHeadings)
	}
	found := false
	for _, h := range s.RequiredHeadings {
		if h == "Tone and tropes" {
			found = true
		}
	}
	if !found {
		t.Fatalf("%v", s.RequiredHeadings)
	}
}

func TestFormatChapterHeadings(t *testing.T) {
	got := FormatChapterHeadings([]string{"A", "B"})
	if !strings.Contains(got, "## A\n") || !strings.Contains(got, "## B\n") {
		t.Fatal(got)
	}
}
