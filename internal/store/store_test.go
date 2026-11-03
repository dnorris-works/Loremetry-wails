package store

import (
	"path/filepath"
	"testing"

	appdb "loremetry/internal/db"
)

// ensureCatalogDB opens a migrated DB and sets catalogDB for tests that don't
// need a full Store but still depend on the analysis_catalog table.
func ensureCatalogDB(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "loremetry.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatalf("open catalog db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	SetCatalogDB(conn)
}

func testStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "loremetry.db")
	conn, err := appdb.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	SetCatalogDB(conn)
	uid, err := appdb.LocalUserID(conn)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	return &Store{DB: conn, UserID: uid}
}

func TestSeriesStoryChapterRoundTrip(t *testing.T) {
	s := testStore(t)

	series, err := s.CreateSeries(SeriesInput{Name: "Cycle", Description: "d", Summary: "s"})
	if err != nil {
		t.Fatalf("create series: %v", err)
	}
	list, err := s.ListSeries()
	if err != nil || len(list) != 1 || list[0].Name != "Cycle" {
		t.Fatalf("list series: %+v %v", list, err)
	}

	story, err := s.CreateStory(StoryInput{Name: "Book 1", SeriesID: series.ID})
	if err != nil {
		t.Fatalf("create story: %v", err)
	}
	name := "ch1.md"
	body := "Once upon a time"
	ch, err := s.CreateChapter(story.ID, ChapterInput{FileName: &name, TextContent: &body})
	if err != nil {
		t.Fatalf("create chapter: %v", err)
	}
	got, err := s.GetChapter(ch.ID)
	if err != nil || got.FileName != "ch1.md" || got.TextContent == nil || *got.TextContent != body {
		t.Fatalf("get chapter: %+v %v", got, err)
	}

	types, err := s.ListDocumentTypes()
	if err != nil || len(types) < 6 {
		t.Fatalf("document types: %+v %v", types, err)
	}

	bible, err := s.CreateSeriesDoc(series.ID, SeriesDocInput{Category: "bible", FileName: "Bible", TextContent: "rules"})
	if err != nil {
		t.Fatalf("series doc: %v", err)
	}
	gotBible, err := s.GetSeriesDoc(bible.ID)
	if err != nil || gotBible.TextContent != "rules" {
		t.Fatalf("get series doc: %+v %v", gotBible, err)
	}

	act, err := s.CreateAct(story.ID, ActInput{Title: "Act 1", Summary: "start"})
	if err != nil {
		t.Fatalf("act: %v", err)
	}
	gotAct, err := s.GetAct(act.ID)
	if err != nil || gotAct.Title != "Act 1" {
		t.Fatalf("get act: %+v %v", gotAct, err)
	}

	sc, err := s.CreateCharacter(CharacterInput{SeriesID: series.ID, CharacterName: "Ava"})
	if err != nil {
		t.Fatalf("series character: %v", err)
	}
	chars, err := s.ListSeriesCharacters(series.ID)
	if err != nil || len(chars) != 1 || chars[0].ID != sc.ID {
		t.Fatalf("list series characters: %+v %v", chars, err)
	}

	if _, err := s.ReorderStories(series.ID, IDList{IDs: []int64{story.ID}}); err != nil {
		t.Fatalf("reorder stories: %v", err)
	}

	tables, err := s.AdminListTables()
	if err != nil || len(tables.Tables) == 0 {
		t.Fatalf("admin tables: %+v %v", tables, err)
	}

	sqlRes, err := s.AdminExecSQL("SELECT name FROM series LIMIT 1000")
	if err != nil || len(sqlRes.Rows) != 1 || sqlRes.Rows[0][0] != "Cycle" {
		t.Fatalf("admin sql: %+v %v", sqlRes, err)
	}

	if _, err := s.DeleteSeriesDoc(bible.ID); err != nil {
		t.Fatalf("delete series doc: %v", err)
	}
	if _, err := s.GetSeriesDoc(bible.ID); err == nil {
		t.Fatal("series doc still present")
	}
	if _, err := s.DeleteAct(act.ID); err != nil {
		t.Fatalf("delete act: %v", err)
	}
}

func TestLocalUserSeeded(t *testing.T) {
	s := testStore(t)
	sess, err := s.Session()
	if err != nil || !sess.Authenticated || !sess.BreakGlass {
		t.Fatalf("session: %+v %v", sess, err)
	}
}
