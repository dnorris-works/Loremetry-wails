package db

import (
	"path/filepath"
	"testing"
)

func TestSeedAnalysisCatalog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "loremetry-app.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM analysis_catalog`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 55 {
		t.Fatalf("catalog count %d want 55", count)
	}

	var instruction, needs string
	err = conn.QueryRow(`SELECT chapter_instruction, needs FROM analysis_catalog WHERE id = 'pov_discipline'`).Scan(&instruction, &needs)
	if err != nil {
		t.Fatal(err)
	}
	if instruction == "" {
		t.Fatal("pov_discipline chapter_instruction empty")
	}
	if needs != `["manuscript","characters"]` {
		t.Fatalf("pov_discipline needs %q", needs)
	}

	var rules string
	err = conn.QueryRow(`SELECT value FROM app_settings WHERE key = 'analysis_source_rules'`).Scan(&rules)
	if err != nil || rules == "" {
		t.Fatalf("analysis_source_rules: %v %q", err, rules)
	}

	checks := map[string]string{
		"chapter_summaries":            `["manuscript","characters"]`,
		"continuity_check":             `["manuscript","characters","locations"]`,
		"ai_beta_reader":               `["manuscript","characters","locations"]`,
		"thematic_throughline":         `["manuscript","themes"]`,
		"recurring_motif_theme_series": `["bible","themes"]`,
	}
	for id, want := range checks {
		var got string
		if err := conn.QueryRow(`SELECT needs FROM analysis_catalog WHERE id = ?`, id).Scan(&got); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if got != want {
			t.Fatalf("%s needs %q want %q", id, got, want)
		}
	}

	var localRunner, sourceInjection string
	err = conn.QueryRow(`SELECT local_runner, source_injection FROM analysis_catalog WHERE id = 'line_polish'`).Scan(&localRunner, &sourceInjection)
	if err != nil {
		t.Fatal(err)
	}
	if localRunner != "line_polish" {
		t.Fatalf("line_polish local_runner %q", localRunner)
	}
	if sourceInjection != "all" {
		t.Fatalf("line_polish source_injection %q", sourceInjection)
	}

	err = conn.QueryRow(`SELECT source_injection FROM analysis_catalog WHERE id = 'pov_discipline'`).Scan(&sourceInjection)
	if err != nil {
		t.Fatal(err)
	}
	if sourceInjection != "selective" {
		t.Fatalf("pov_discipline source_injection %q", sourceInjection)
	}
}
