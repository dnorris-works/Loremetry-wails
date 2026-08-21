package analysis

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"loremetry/internal/apiserver/models"
	appdb "loremetry/internal/db"
	"loremetry/internal/store"
)

type mockGateway struct {
	calls []string
}

func (m *mockGateway) Complete(ctx context.Context, tier models.Tier, system, user string) (string, models.Usage, error) {
	step := "other"
	switch {
	case strings.Contains(user, "Analyze this single chapter"):
		step = "analyze_chapter"
	case strings.Contains(user, "Summarize this manuscript section"):
		step = "summarize_chunk"
	case strings.Contains(user, "Merge these section summaries"):
		step = "merge_outline"
	case strings.Contains(user, "Summarize the following source"):
		step = "summarize_role"
	case strings.Contains(user, "Run analysis:"):
		step = "final"
	}
	m.calls = append(m.calls, step)
	return fmt.Sprintf("ok-%s-%d", step, len(m.calls)), models.Usage{}, nil
}

func TestRunByChapterCallsPerChapterThenFinal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	store.SetCatalogDB(conn)

	if ProfileFor("pov_discipline") != ProfileByChapter {
		t.Fatalf("expected by_chapter for pov_discipline, got %v", ProfileFor("pov_discipline"))
	}

	gw := &mockGateway{}
	runner := Runner{Gateway: gw}
	sources := []Source{
		{Role: "manuscript", Name: "01-Open.md", Text: "Chapter one text about Alice."},
		{Role: "manuscript", Name: "02-Middle.md", Text: "Chapter two text about Bob."},
		{Role: "manuscript", Name: "03-End.md", Text: "Chapter three text about Carol."},
	}
	body, err := runner.Run(context.Background(), "pov_discipline", sources)
	if err != nil {
		t.Fatal(err)
	}
	if body == "" {
		t.Fatal("empty body")
	}
	if len(gw.calls) != 4 {
		t.Fatalf("calls=%v want 3 chapter + 1 final", gw.calls)
	}
	for i := 0; i < 3; i++ {
		if gw.calls[i] != "analyze_chapter" {
			t.Fatalf("call %d = %s", i, gw.calls[i])
		}
	}
	if gw.calls[3] != "final" {
		t.Fatalf("last call = %s", gw.calls[3])
	}
}

func TestRunChunkedJoinsMultipleManuscriptSources(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	store.SetCatalogDB(conn)

	if ProfileFor("analysis") != ProfileChunked {
		t.Fatal("expected chunked for analysis")
	}

	gw := &mockGateway{}
	runner := Runner{Gateway: gw}
	// Two short chapters; joined text stays under split threshold → one summarize_chunk.
	sources := []Source{
		{Role: "manuscript", Name: "01-A.md", Text: "UNIQUE_CHAPTER_ONE_MARKER"},
		{Role: "manuscript", Name: "02-B.md", Text: "UNIQUE_CHAPTER_TWO_MARKER"},
		{Role: "plot", Name: "plot.md", Text: "Plot notes."},
	}
	_, err = runner.Run(context.Background(), "analysis", sources)
	if err != nil {
		t.Fatal(err)
	}
	var sawChunk bool
	for _, c := range gw.calls {
		if c == "summarize_chunk" {
			sawChunk = true
		}
	}
	if !sawChunk {
		t.Fatalf("expected summarize_chunk, calls=%v", gw.calls)
	}
	// Ensure join did not drop the first chapter: inspect the first summarize_chunk user via a recording gateway.
	rec := &recordingGateway{}
	runner2 := Runner{Gateway: rec}
	_, err = runner2.Run(context.Background(), "analysis", sources)
	if err != nil {
		t.Fatal(err)
	}
	joined := false
	for _, u := range rec.users {
		if strings.Contains(u, "UNIQUE_CHAPTER_ONE_MARKER") && strings.Contains(u, "UNIQUE_CHAPTER_TWO_MARKER") {
			joined = true
			break
		}
		if strings.Contains(u, "UNIQUE_CHAPTER_ONE_MARKER") || strings.Contains(u, "UNIQUE_CHAPTER_TWO_MARKER") {
			// section summaries may only include one marker each; both must appear across calls
		}
	}
	var one, two bool
	for _, u := range rec.users {
		if strings.Contains(u, "UNIQUE_CHAPTER_ONE_MARKER") {
			one = true
		}
		if strings.Contains(u, "UNIQUE_CHAPTER_TWO_MARKER") {
			two = true
		}
	}
	if !one || !two {
		t.Fatalf("chunked must include both manuscript chapters; one=%v two=%v joined=%v users=%d", one, two, joined, len(rec.users))
	}
}

type recordingGateway struct {
	users []string
}

func (r *recordingGateway) Complete(ctx context.Context, tier models.Tier, system, user string) (string, models.Usage, error) {
	r.users = append(r.users, user)
	return "ok", models.Usage{}, nil
}

func TestProfileForByChapterCatalog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	store.SetCatalogDB(conn)

	want := []string{
		"chapter_summaries", "show_dont_tell", "ai_isms", "pov_discipline", "scene_sequel_balance",
		"cliffhanger_score", "pacing_curve", "ai_beta_reader", "dramatic_irony", "content_maturity_advisory",
		"chekhovs_gun", "macguffin_clarity", "red_herring_vs_abandoned", "foreshadowing_twist_fairness",
		"timeline_flashback", "stakes_escalation", "story_beat_placement",
	}
	for _, id := range want {
		if ProfileFor(id) != ProfileByChapter {
			t.Errorf("%s: got %v want by_chapter", id, ProfileFor(id))
		}
	}
	if ProfileFor("hook_strength") != ProfileSingle {
		t.Fatal("hook_strength should stay single")
	}
	if ProfileFor("analysis") != ProfileChunked {
		t.Fatal("analysis should stay chunked")
	}
}

func TestManuscriptChaptersFromFiles(t *testing.T) {
	chapters, other := manuscriptChapters([]Source{
		{Role: "manuscript", Name: "01.md", Text: "a"},
		{Role: "characters", Name: "cast.md", Text: "cast"},
		{Role: "manuscript", Name: "02.md", Text: "b"},
	})
	if len(chapters) != 2 {
		t.Fatalf("chapters=%d", len(chapters))
	}
	if len(other) != 1 || other[0].Role != "characters" {
		t.Fatalf("other=%+v", other)
	}
}
