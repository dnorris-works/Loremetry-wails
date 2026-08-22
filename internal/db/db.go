package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schemaVersion = 18

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	dir := filepath.Join(configDir, "loremetry")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create data dir: %w", err)
	}
	// loremetry.db in this folder belongs to the older desktop app. Do not open or migrate it.
	return filepath.Join(dir, "loremetry-app.db"), nil
}

func Open(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	if _, err := conn.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := migrate(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := seedLocalUser(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func migrate(conn *sql.DB) error {
	var version int
	if err := conn.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version >= schemaVersion {
		return nil
	}
	if _, err := conn.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	_, _ = conn.Exec(`ALTER TABLE stories ADD COLUMN series_id INTEGER`)
	_, _ = conn.Exec(`ALTER TABLE stories ADD COLUMN series_sort_order INTEGER`)
	_, _ = conn.Exec(`ALTER TABLE story_documents ADD COLUMN act_id INTEGER`)
	_, _ = conn.Exec(`ALTER TABLE story_documents ADD COLUMN source_rel TEXT`)
	_, _ = conn.Exec(`ALTER TABLE series_bible_documents ADD COLUMN source_rel TEXT`)
	_, _ = conn.Exec(`ALTER TABLE character_profiles ADD COLUMN source_rel TEXT`)
	_, _ = conn.Exec(`ALTER TABLE story_acts ADD COLUMN source_rel TEXT`)
	_, _ = conn.Exec(`ALTER TABLE series ADD COLUMN pen_name TEXT`)
	_, _ = conn.Exec(`ALTER TABLE stories ADD COLUMN pen_name TEXT`)
	_, _ = conn.Exec(`ALTER TABLE analysis_reports ADD COLUMN original TEXT NOT NULL DEFAULT ''`)
	_, _ = conn.Exec(`ALTER TABLE analysis_reports ADD COLUMN proposed TEXT NOT NULL DEFAULT ''`)
	if err := ensureCharacterParents(conn); err != nil {
		return err
	}
	if _, err := conn.Exec(indexSQL); err != nil {
		return fmt.Errorf("apply indexes: %w", err)
	}
	_, _ = conn.Exec(`INSERT OR IGNORE INTO app_settings (key, value) VALUES ('cloud_api_base_url', 'https://api.loremetry.com')`)
	_, err := conn.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, schemaVersion))
	return err
}

func ensureCharacterParents(conn *sql.DB) error {
	rows, err := conn.Query(`PRAGMA table_info(character_profiles)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	storyNotNull := false
	hasSeries := false
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "story_id" && notnull == 1 {
			storyNotNull = true
		}
		if name == "series_id" {
			hasSeries = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if hasSeries && !storyNotNull {
		return nil
	}
	_, err = conn.Exec(`
		CREATE TABLE character_profiles_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			series_id INTEGER REFERENCES series(id) ON DELETE CASCADE,
			story_id INTEGER REFERENCES stories(id) ON DELETE CASCADE,
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
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
		INSERT INTO character_profiles_new (
			id, story_id, character_name, role_in_story, story_importance, pov_character,
			character_want, character_need, arc_status_at_start, arc_status_at_end,
			key_relationships, defining_trait, first_appearance_book, primary_obstacle,
			internal_vs_external, core_wound, age, thematic_resonance, character_voice_notes,
			created_at, updated_at
		)
		SELECT id, story_id, character_name, role_in_story, story_importance, pov_character,
			character_want, character_need, arc_status_at_start, arc_status_at_end,
			key_relationships, defining_trait, first_appearance_book, primary_obstacle,
			internal_vs_external, core_wound, age, thematic_resonance, character_voice_notes,
			created_at, updated_at
		FROM character_profiles;
		DROP TABLE character_profiles;
		ALTER TABLE character_profiles_new RENAME TO character_profiles;
		CREATE INDEX IF NOT EXISTS idx_char_profiles_story ON character_profiles(story_id);
		CREATE INDEX IF NOT EXISTS idx_char_profiles_series ON character_profiles(series_id);
	`)
	return err
}

func seedLocalUser(conn *sql.DB) error {
	_, err := conn.Exec(`
		INSERT INTO users (email, first_name, last_name, created_at, updated_at)
		SELECT 'local@loremetry', 'Local', 'User', datetime('now'), datetime('now')
		WHERE NOT EXISTS (SELECT 1 FROM users LIMIT 1)
	`)
	return err
}

func LocalUserID(conn *sql.DB) (int64, error) {
	var id int64
	err := conn.QueryRow(`SELECT id FROM users ORDER BY id LIMIT 1`).Scan(&id)
	return id, err
}
