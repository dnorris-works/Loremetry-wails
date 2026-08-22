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

DROP TABLE IF EXISTS analysis_catalog;
CREATE TABLE analysis_catalog (
    id          TEXT PRIMARY KEY,
    group_id    TEXT NOT NULL,
    group_label TEXT NOT NULL,
    label       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    usage       TEXT NOT NULL DEFAULT '',
    needs       TEXT NOT NULL DEFAULT '[]',
    depends_on  TEXT NOT NULL DEFAULT '[]',
    uses_ai     INTEGER NOT NULL DEFAULT 0,
    uses_merge  INTEGER NOT NULL DEFAULT 0,
    ai_profile  TEXT NOT NULL DEFAULT 'single',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    chapter_instruction TEXT NOT NULL DEFAULT '',
    chapter_headings    TEXT NOT NULL DEFAULT '[]',
    final_instruction   TEXT NOT NULL DEFAULT '',
    source_injection    TEXT NOT NULL DEFAULT 'selective',
    local_runner        TEXT NOT NULL DEFAULT '',
    local_config        TEXT NOT NULL DEFAULT '{}'
);

-- Group: KDP / Wide
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('chapter_summaries', 'kdp-wide', 'KDP / Wide', 'Chapter Plot Summary', 'Real plot summary per chapter: who acts, what happens, stakes, and open loops.', 'Optional reference for you—not required by other analyses. Each chapter is summarized with character profiles, then assembled into one document.', '["manuscript","characters"]', '[]', 1, 0, 'by_chapter', 1);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('analysis', 'kdp-wide', 'KDP / Wide', 'KDP Analysis', 'Genre, Kindle and paperback categories, print BISAC, seven keywords, and ready-to-paste KDP metadata.', 'Copy categories, BISAC, and keywords into KDP. Save this report before you change the manuscript.', '["manuscript","plot"]', '["mi_search_terms"]', 1, 0, 'by_chapter', 2);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('genre_analysis', 'kdp-wide', 'KDP / Wide', 'Genre Analysis', 'Industry genre classification, ranking, comparable titles, and reader demographic.', 'Use the top genre and comps when choosing categories and writing your blurb.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 3);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('genre_ranking', 'kdp-wide', 'KDP / Wide', 'Genre Ranking', 'Score the manuscript against known genres independently.', 'Compare scores across genres; pick the strongest fit for metadata and positioning.', '["manuscript"]', '["genre_analysis"]', 1, 0, 'by_chapter', 4);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('kdp_categories', 'kdp-wide', 'KDP / Wide', 'KDP Categories', 'Best-fit Amazon browse paths for Kindle and/or paperback.', 'Enter the suggested Kindle and paperback browse paths in KDP category fields.', '["manuscript"]', '["genre_analysis"]', 1, 0, 'by_chapter', 5);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('kdp_keywords', 'kdp-wide', 'KDP / Wide', 'KDP Keywords', 'Seven KDP keyword strings (50 characters or fewer).', 'Paste each string into KDP''s seven keyword slots (50 characters or fewer).', '["manuscript","plot"]', '["genre_analysis"]', 1, 0, 'by_chapter', 6);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('mi_search_terms', 'kdp-wide', 'KDP / Wide', 'Search Terms', 'Short Amazon-style phrases for competition research.', 'Use these phrases in Amazon search and competition tools to size the niche.', '["manuscript"]', '["analysis"]', 1, 0, 'by_chapter', 7);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('keyword_search', 'kdp-wide', 'KDP / Wide', 'Keyword Search Results', 'Amazon keyword volume and competition data.', 'Favor high-volume, moderate-competition terms when refining KDP keywords.', '["manuscript"]', '["analysis"]', 1, 0, 'by_chapter', 8);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('competition_report', 'kdp-wide', 'KDP / Wide', 'Competition Analysis', 'Market landscape: niche competitiveness, comps, pricing, and viability.', 'Weigh niche density and comps before locking genre, price, and series strategy.', '["manuscript"]', '["mi_search_terms"]', 1, 0, 'by_chapter', 9);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('review_mining', 'kdp-wide', 'KDP / Wide', 'Reader Review Intelligence', 'What readers love and hate in competitor reviews, plus positioning notes.', 'Steal reader language for blurbs and fix pain points they complain about in comps.', '["manuscript"]', '["mi_search_terms"]', 1, 0, 'by_chapter', 10);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('author_analysis', 'kdp-wide', 'KDP / Wide', 'Competitor Author Analysis', 'Competitor catalogs: cadence, pricing, series vs standalone, and takeaways.', 'Benchmark release cadence, pricing, and series length against nearby competitors.', '["manuscript"]', '["mi_search_terms"]', 1, 0, 'by_chapter', 11);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('wide_analysis', 'kdp-wide', 'KDP / Wide', 'Wide Analysis', 'BISAC, discovery keywords, Google SEO, content advisory, and wide paste metadata.', 'Apply BISAC, discovery keywords, and content notes across wide stores and aggregators.', '["manuscript","plot"]', '[]', 1, 0, 'by_chapter', 12);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('bisac_classification', 'kdp-wide', 'KDP / Wide', 'BISAC Classification', 'BISAC subject codes for Ingram, Apple, Kobo, and aggregators.', 'Enter the recommended BISAC codes in Ingram, Apple, Kobo, or your aggregator.', '["manuscript"]', '["genre_analysis"]', 1, 0, 'by_chapter', 13);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('discovery_keywords', 'kdp-wide', 'KDP / Wide', 'Discovery Keywords', 'Ten wide-store SEO phrases.', 'Add these phrases to wide-store keyword and SEO fields.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 14);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('google_keyword_search', 'kdp-wide', 'KDP / Wide', 'Google Keyword Search', 'Google search volume for wide-store phrases.', 'Prioritize phrases with real search volume when filling wide discovery keywords.', '["manuscript"]', '["discovery_keywords"]', 1, 0, 'by_chapter', 15);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('content_maturity_advisory', 'kdp-wide', 'KDP / Wide', 'Content & Maturity Advisory', 'Heat level, content warnings, and age guidance for wide stores.', 'Match store maturity settings and content warnings to the advisory below.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 16);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('wide_metadata_paste', 'kdp-wide', 'KDP / Wide', 'Wide Metadata Paste Sheet', 'Copy-ready wide metadata for aggregators and stores.', 'Copy each block into the matching aggregator or store metadata field.', '["plot"]', '["bisac_classification","discovery_keywords"]', 1, 0, 'single', 17);

-- Group: Craft — Prose
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('show_dont_tell', 'prose', 'Craft — Prose', 'Show Don''t Tell', 'Flags passages that tell emotion or judgment instead of showing it.', 'Rewrite flagged telling passages to show emotion through action and sensory detail.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 18);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('ai_isms', 'prose', 'Craft — Prose', 'AI-isms', 'Flags prose that often reads as machine-generated.', 'Revise flagged lines so they sound like your voice, not generic machine prose.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 19);

-- Group: Craft — Structure & Pacing
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('story_beat_placement', 'structure', 'Craft — Structure & Pacing', 'Story Beat Placement', 'Beat timing against common story frameworks.', 'Move, add, or cut beats so structural milestones land where the framework expects.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 20);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('scene_sequel_balance', 'structure', 'Craft — Structure & Pacing', 'Scene/Sequel Balance', 'Action vs reflective passages and where momentum stalls.', 'Add action or reflection where the report shows momentum stalling.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 21);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('pov_discipline', 'structure', 'Craft — Structure & Pacing', 'POV Discipline', 'POV shifts, head-hopping, and information leaks.', 'Fix head-hops and information leaks so each scene stays in the intended POV.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 22);

-- Group: Craft — Plot & Continuity
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('continuity_check', 'plot', 'Craft — Plot & Continuity', 'Continuity Check', 'Finds contradicted facts in a manuscript or across a series.', 'Fix contradicted facts in the manuscript (or series bible) before the next draft pass.', '["manuscript","characters","locations"]', '[]', 1, 0, 'by_chapter', 23);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('chekhovs_gun', 'plot', 'Craft — Plot & Continuity', 'Chekhov''s Gun', 'Early setups (objects, skills, promises) and whether they pay off.', 'Pay off unused setups or cut them so early mentions earn their keep.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 24);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('red_herring_vs_abandoned', 'plot', 'Craft — Plot & Continuity', 'Red Herring vs Abandoned', 'Separates intentional misdirection from dropped plot threads.', 'Keep intentional misdirection; restore or remove dropped threads.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 25);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('foreshadowing_twist_fairness', 'plot', 'Craft — Plot & Continuity', 'Foreshadowing & Twist Fairness', 'Whether foreshadowing is fair and twists feel earned.', 'Add fair clues where twists feel cheap; trim over-telegraphing where they feel obvious.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 26);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('macguffin_clarity', 'plot', 'Craft — Plot & Continuity', 'MacGuffin Clarity', 'Whether the driving object or goal is clear and motivates action.', 'Clarify the driving object or goal if characters (or readers) lose track of why it matters.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 27);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('timeline_flashback', 'plot', 'Craft — Plot & Continuity', 'Timeline / Flashback', 'Whether timeline shifts and flashbacks clarify or confuse.', 'Clarify or cut flashbacks that confuse chronology rather than deepen stakes.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 28);

-- Group: Craft — Character & Theme
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('want_vs_need', 'character', 'Craft — Character & Theme', 'Want vs Need', 'External want vs internal need for major characters.', 'Align scenes so external want and internal need pull characters through the arc.', '["manuscript","characters"]', '[]', 1, 0, 'by_chapter', 29);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('thematic_throughline', 'character', 'Craft — Character & Theme', 'Thematic Throughline', 'How the central theme holds across scenes and arcs.', 'Reinforce or prune scenes that drift from the central theme.', '["manuscript","bible"]', '[]', 1, 0, 'by_chapter', 30);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('mirror_foil_character', 'character', 'Craft — Character & Theme', 'Mirror/Foil Characters', 'Reflect and contrast pairings and what they do for theme.', 'Use mirror/foil pairings to sharpen theme; cut pairings that do no thematic work.', '["manuscript","characters"]', '[]', 1, 0, 'by_chapter', 31);

-- Group: Craft — Reader Engagement
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('zeigarnik_analysis', 'engagement', 'Craft — Reader Engagement', 'Zeigarnik Effect', 'Heuristic scan for open loops, cliffhangers, and unresolved threads. No AI.', 'Use Compare to review flagged chapter endings; strengthen weak open loops and leave intentional closures alone.', '["manuscript"]', '[]', 0, 1, 'single', 32);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('dramatic_irony', 'engagement', 'Craft — Reader Engagement', 'Dramatic Irony', 'Moments the reader knows more than the characters.', 'Lean into reader-knows-more moments where tension or humor pays off.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 33);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('stakes_escalation', 'engagement', 'Craft — Reader Engagement', 'Stakes Escalation', 'How stakes rise, plateau, or reverse across the arc.', 'Raise or reset stakes where the arc plateaus; cut reversals that undercut tension.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 34);

-- Group: Craft — Series
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('cross_book_setup_payoff', 'series', 'Craft — Series', 'Cross-Book Setup/Payoff', 'Series setups planted in earlier books that should pay off later.', 'Track series setups that still need payoff in a later book or epilogue.', '["bible","characters"]', '[]', 1, 0, 'two_step', 35);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('series_pacing_comparator', 'series', 'Craft — Series', 'Series Pacing Comparator', 'Pacing compared across books in a series.', 'Adjust book length and beat density so series pacing feels consistent.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 36);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('recurring_motif_theme_series', 'series', 'Craft — Series', 'Recurring Motif/Theme (Series)', 'Motifs and themes across books — cohesion vs contradiction.', 'Keep motifs coherent across books; resolve contradictions in theme or symbol.', '["bible"]', '[]', 1, 0, 'two_step', 37);

-- Group: Publish
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('blurb_builder', 'publish', 'Publish', 'Blurb Builder', 'Amazon, back-cover, and BookBub description variants.', 'Pick a variant for Amazon, back cover, or BookBub and paste it into that channel.', '["manuscript","bible","blurb"]', '[]', 1, 0, 'by_chapter', 38);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('print_production', 'publish', 'Publish', 'Print Production', 'Page count, trim, spine, and print checklists from word count. No AI.', 'Use word count and spine estimates when choosing trim and ordering a KDP/Ingram cover template.', '["manuscript"]', '[]', 0, 0, 'single', 39);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('ai_beta_reader', 'publish', 'Publish', 'AI Beta Reader', 'Chapter-by-chapter reader reactions, engagement, and put-down risk.', 'Prioritize chapters with high put-down risk; revise hooks and endings there first.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 40);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('cliffhanger_score', 'publish', 'Publish', 'Cliffhanger Score', 'How hard each chapter ending pulls into the next.', 'Strengthen weak chapter endings so readers turn the page.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 41);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('hook_strength', 'publish', 'Publish', 'Hook Strength', 'Whether a browsing reader would keep going past page one.', 'Rewrite the opening if a browsing reader would stop before page two.', '["manuscript"]', '[]', 1, 0, 'single', 42);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('pacing_curve', 'publish', 'Publish', 'Pacing Curve', 'Per-chapter pace and drag-risk scores.', 'Trim or split chapters marked as drag risk; keep high-pace chapters intact.', '["manuscript"]', '[]', 1, 0, 'by_chapter', 43);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('line_polish', 'publish', 'Publish', 'Line-level Polish', 'Filter words, echoes, adverbs, and rough passives. No AI.', 'Search the manuscript for listed filter words and echoes; cut or replace the worst repeats.', '["manuscript"]', '[]', 0, 0, 'single', 44);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('vellum_prep', 'publish', 'Publish', 'Vellum & Atticus Prep', 'Clean markdown manuscript for Vellum or Atticus import. No AI.', 'Copy or export this markdown into Vellum or Atticus as a new book import.', '["manuscript"]', '[]', 0, 0, 'single', 45);

-- Group: Craft — Metrics
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('readability_score', 'metrics', 'Craft — Metrics', 'Readability Score', 'Flesch readability and grade-level metrics per chapter.', 'Review per-chapter Flesch scores and adjust sentence complexity where grade level is too high.', '["manuscript"]', '[]', 0, 0, 'single', 46);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('sentence_length_variation', 'metrics', 'Craft — Metrics', 'Sentence Length Variation', 'Sentence length variance and rhythm analysis per chapter.', 'Vary sentence lengths in chapters flagged as monotonous to improve rhythm.', '["manuscript"]', '[]', 0, 0, 'single', 47);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('chapter_balance', 'metrics', 'Craft — Metrics', 'Chapter Balance', 'Word-count balance and outlier detection across chapters.', 'Consider splitting long chapters or expanding short ones to even the reading pace.', '["manuscript"]', '[]', 0, 0, 'single', 48);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('dialogue_ratio', 'metrics', 'Craft — Metrics', 'Dialogue Ratio', 'Dialogue-to-narration ratio per chapter and overall.', 'Add dialogue to narration-heavy chapters or vice versa to improve reader engagement.', '["manuscript"]', '[]', 0, 0, 'single', 49);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('dialogue_tag_audit', 'metrics', 'Craft — Metrics', 'Dialogue Tag Audit', 'Dialogue tag frequency and overuse detection.', 'Replace overused tags with action beats or cut redundant attribution.', '["manuscript"]', '[]', 0, 0, 'single', 50);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('passive_voice', 'metrics', 'Craft — Metrics', 'Passive Voice Density', 'Passive voice density per chapter.', 'Rewrite passive constructions to active voice where clarity and energy improve.', '["manuscript"]', '[]', 0, 0, 'single', 51);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('sticky_sentences', 'metrics', 'Craft — Metrics', 'Sticky Sentences', 'Glue-word density and sticky sentence detection.', 'Rewrite sticky sentences by replacing glue words with concrete nouns and verbs.', '["manuscript"]', '[]', 0, 0, 'single', 52);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('repeated_phrases', 'metrics', 'Craft — Metrics', 'Repeated Phrase Finder', 'Repeated n-gram detection across the manuscript.', 'Vary repeated phrases or find synonyms to reduce reader fatigue.', '["manuscript"]', '[]', 0, 0, 'single', 53);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('paragraph_length', 'metrics', 'Craft — Metrics', 'Paragraph Length Stats', 'Paragraph length statistics per chapter.', 'Break up long paragraphs or combine short ones for better visual rhythm.', '["manuscript"]', '[]', 0, 0, 'single', 54);
INSERT INTO analysis_catalog (id, group_id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order) VALUES
  ('opening_closing', 'metrics', 'Craft — Metrics', 'Opening/Closing Strength', 'Opening and closing sentence strength heuristics per chapter.', 'Strengthen weak chapter openings with hooks and closings with forward momentum.', '["manuscript"]', '[]', 0, 0, 'single', 55);

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

CREATE TABLE IF NOT EXISTS analysis_run_stats (
    analysis_id TEXT PRIMARY KEY,
    run_count   INTEGER NOT NULL DEFAULT 0,
    last_ms     INTEGER NOT NULL DEFAULT 0,
    avg_ms      INTEGER NOT NULL DEFAULT 0,
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS chapter_descriptions (
    project_path TEXT NOT NULL,
    chapter_rel  TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    updated_at   TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (project_path, chapter_rel)
);
`
