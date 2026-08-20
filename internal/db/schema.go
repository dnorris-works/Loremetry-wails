package db

const indexSQL = `
CREATE INDEX IF NOT EXISTS idx_series_user_updated ON series(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_stories_user_updated ON stories(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_stories_series_sort ON stories(series_id, series_sort_order);
CREATE INDEX IF NOT EXISTS idx_story_docs_story_sort ON story_documents(story_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_story_acts_story ON story_acts(story_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_char_profiles_story ON character_profiles(story_id);
CREATE INDEX IF NOT EXISTS idx_char_profiles_series ON character_profiles(series_id);
CREATE INDEX IF NOT EXISTS idx_series_bible_sort ON series_bible_documents(series_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_analysis_reports_user ON analysis_reports(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analysis_results_user ON analysis_results(user_id, project_path, updated_at DESC);
`

const schemaSQL = `
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    clerk_id    TEXT UNIQUE,
    email       TEXT NOT NULL UNIQUE,
    first_name  TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS series (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    summary     TEXT,
    pen_name    TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS stories (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id           INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    series_id         INTEGER REFERENCES series(id) ON DELETE SET NULL,
    series_sort_order INTEGER,
    name              TEXT NOT NULL,
    description       TEXT,
    pen_name          TEXT,
    created_at        TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at        TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS story_acts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id   INTEGER NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    summary    TEXT,
    sort_order INTEGER,
    source_rel TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS story_documents (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id       INTEGER NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    kind           TEXT NOT NULL,
    file_name      TEXT NOT NULL,
    mime_type      TEXT NOT NULL,
    text_content   TEXT,
    binary_content BLOB,
    sort_order     INTEGER,
    act_id         INTEGER REFERENCES story_acts(id) ON DELETE SET NULL,
    source_rel     TEXT,
    created_at     TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(story_id, kind, file_name)
);

CREATE TABLE IF NOT EXISTS character_profiles (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    series_id             INTEGER REFERENCES series(id) ON DELETE CASCADE,
    story_id              INTEGER REFERENCES stories(id) ON DELETE CASCADE,
    character_name        TEXT NOT NULL,
    role_in_story         TEXT,
    story_importance      TEXT,
    pov_character         INTEGER NOT NULL DEFAULT 0,
    character_want        TEXT,
    character_need        TEXT,
    arc_status_at_start   TEXT,
    arc_status_at_end     TEXT,
    key_relationships     TEXT,
    defining_trait        TEXT,
    first_appearance_book INTEGER,
    primary_obstacle      TEXT,
    internal_vs_external  TEXT,
    core_wound            TEXT,
    age                   INTEGER,
    thematic_resonance    TEXT,
    character_voice_notes TEXT,
    source_rel            TEXT,
    created_at            TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at            TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS series_bible_documents (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    series_id      INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    category       TEXT NOT NULL,
    file_name      TEXT NOT NULL,
    mime_type      TEXT NOT NULL,
    text_content   TEXT,
    binary_content BLOB,
    sort_order     INTEGER,
    source_rel     TEXT,
    created_at     TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(series_id, category, file_name)
);

DROP TABLE IF EXISTS document_types;
CREATE TABLE document_types (
    code               TEXT PRIMARY KEY,
    display_name       TEXT NOT NULL,
    sort_order         INTEGER,
    applies_to_story   INTEGER NOT NULL DEFAULT 0,
    applies_to_series  INTEGER NOT NULL DEFAULT 0,
    active             INTEGER NOT NULL DEFAULT 1,
    editor             TEXT NOT NULL,
    storage            TEXT NOT NULL
);

INSERT INTO document_types (code, display_name, sort_order, applies_to_story, applies_to_series, active, editor, storage) VALUES
  ('bible',      'Bible',      0, 1, 1, 1, 'prose',      'doc'),
  ('location',   'Location',   1, 1, 1, 1, 'prose',      'doc'),
  ('research',   'Research',   2, 1, 1, 1, 'prose',      'doc'),
  ('character',  'Character',  3, 1, 1, 1, 'character',  'character'),
  ('act',        'Act',        4, 1, 0, 1, 'act',        'act'),
  ('chapter',    'Chapter',    5, 1, 0, 1, 'prose',      'doc');

CREATE TABLE IF NOT EXISTS user_settings (
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        TEXT NOT NULL,
    value      TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, key)
);

CREATE TABLE IF NOT EXISTS app_settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS folder_links (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    scope    TEXT NOT NULL,
    owner_id INTEGER NOT NULL,
    kind     TEXT NOT NULL,
    path     TEXT NOT NULL,
    UNIQUE(scope, owner_id, kind)
);

CREATE TABLE IF NOT EXISTS header_overrides (
    project_path TEXT NOT NULL,
    kind         TEXT NOT NULL,
    path         TEXT NOT NULL,
    PRIMARY KEY (project_path, kind)
);

CREATE TABLE IF NOT EXISTS disk_hashes (
    root TEXT NOT NULL,
    rel  TEXT NOT NULL,
    hash TEXT NOT NULL,
    PRIMARY KEY (root, rel)
);

CREATE TABLE IF NOT EXISTS analysis_reports (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    analysis_id   TEXT NOT NULL,
    analysis_label TEXT NOT NULL,
    project_path  TEXT NOT NULL,
    uses_ai       INTEGER NOT NULL DEFAULT 0,
    body          TEXT NOT NULL,
    original      TEXT NOT NULL DEFAULT '',
    proposed      TEXT NOT NULL DEFAULT '',
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS analysis_results (
    user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_path TEXT NOT NULL,
    analysis_id  TEXT NOT NULL,
    label        TEXT NOT NULL,
    data_json    TEXT NOT NULL DEFAULT '{}',
    markdown     TEXT NOT NULL DEFAULT '',
    original     TEXT NOT NULL DEFAULT '',
    proposed     TEXT NOT NULL DEFAULT '',
    updated_at   TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (user_id, project_path, analysis_id)
);
`
