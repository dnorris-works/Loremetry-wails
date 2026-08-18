package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestDefaultPathUsesAppDatabase(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "loremetry-app.db" {
		t.Fatalf("got %s", path)
	}
}

func TestOpenCreatesAppTables(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	opened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = opened.Close() })

	want := []string{
		"users", "series", "stories", "story_acts", "story_documents", "character_profiles",
		"series_bible_documents", "document_types", "user_settings", "app_settings", "disk_hashes",
		"analysis_reports",
	}
	for _, table := range want {
		var name string
		err := opened.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("missing table %s: %v", table, err)
		}
	}
}

func TestMigrateOldCharacterProfiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, first_name TEXT, last_name TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE series (id INTEGER PRIMARY KEY, user_id INTEGER, name TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE stories (id INTEGER PRIMARY KEY, user_id INTEGER, series_id INTEGER, name TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE character_profiles (
			id INTEGER PRIMARY KEY,
			story_id INTEGER NOT NULL,
			character_name TEXT NOT NULL,
			role_in_story TEXT,
			story_importance TEXT,
			pov_character INTEGER NOT NULL DEFAULT 0,
			character_want TEXT,
			character_need TEXT,
			arc_status_at_start TEXT,
			arc_status_at_end TEXT,
			key_relationships TEXT,
			defining_trait TEXT,
			first_appearance_book INTEGER,
			primary_obstacle TEXT,
			internal_vs_external TEXT,
			core_wound TEXT,
			age INTEGER,
			thematic_resonance TEXT,
			character_voice_notes TEXT,
			created_at TEXT,
			updated_at TEXT
		);
		CREATE TABLE story_documents (
			id INTEGER PRIMARY KEY,
			story_id INTEGER NOT NULL,
			kind TEXT NOT NULL,
			file_name TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			text_content TEXT,
			binary_content BLOB,
			sort_order INTEGER,
			act INTEGER,
			created_at TEXT
		);
		PRAGMA user_version = 1;
	`)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()

	opened, err := Open(path)
	if err != nil {
		t.Fatalf("migrate old db: %v", err)
	}
	t.Cleanup(func() { _ = opened.Close() })

	var n int
	err = opened.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('character_profiles') WHERE name = 'series_id'`).Scan(&n)
	if err != nil || n != 1 {
		t.Fatalf("series_id missing after migrate: %v count=%d", err, n)
	}
}
