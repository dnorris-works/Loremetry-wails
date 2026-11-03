package analysis

import (
	"strings"
	"testing"
)

func TestFilterStoryElementsForChapter(t *testing.T) {
	other := []Source{
		{Role: "characters", Name: "Maya.md", Text: "# Maya\n\nProtagonist."},
		{Role: "characters", Name: "Zoe.md", Text: "# Zoe\n\nRoommate."},
		{Role: "plot", Name: "outline.md", Text: "Plot notes"},
	}
	ch := Source{Role: "manuscript", Name: "01.md", Text: "Maya opened the door."}
	got := FilterStoryElementsForChapter(other, ch)
	if len(got) != 2 {
		t.Fatalf("got %d sources: %+v", len(got), got)
	}
	if got[0].Role != "characters" || got[0].Name != "Maya.md" {
		t.Fatalf("expected Maya profile, got %+v", got[0])
	}
	if got[1].Role != "plot" {
		t.Fatalf("expected plot kept, got %+v", got[1])
	}
}

func TestFilterStoryElementsIncludesBothNames(t *testing.T) {
	other := []Source{
		{Role: "characters", Name: "Maya.md", Text: "# Maya"},
		{Role: "characters", Name: "Zoe.md", Text: "# Zoe"},
	}
	ch := Source{Role: "manuscript", Name: "01.md", Text: "Maya and Zoe talk."}
	got := FilterStoryElementsForChapter(other, ch)
	if len(got) != 2 {
		t.Fatalf("expected both profiles, got %+v", got)
	}
}

func TestContainsWord(t *testing.T) {
	if !containsWord("maya opened the door", "maya") {
		t.Fatal("expected maya match")
	}
	if containsWord("maya opened", "may") {
		t.Fatal("should not match substring")
	}
	if !containsWord(strings.ToLower("Zoe smiled."), "zoe") {
		t.Fatal("expected zoe match")
	}
}
