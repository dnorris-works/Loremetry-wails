package db

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed catalog/*.json
var catalogFS embed.FS

const analysisSourceRules = "Treat characters, locations, and themes sources as authoritative for names, relationships, settings, and motifs. Do not invent details that contradict those sources or are unsupported by the manuscript."

type catalogEntry struct {
	ID                  string          `json:"id"`
	GroupID             string          `json:"group_id"`
	GroupLabel          string          `json:"group_label"`
	Label               string          `json:"label"`
	Description         string          `json:"description"`
	Usage               string          `json:"usage"`
	Needs               []string        `json:"needs"`
	DependsOn           []string        `json:"depends_on"`
	UsesAI              bool            `json:"uses_ai"`
	UsesMerge           bool            `json:"uses_merge"`
	AIProfile           string          `json:"ai_profile"`
	SortOrder           int             `json:"sort_order"`
	ChapterInstruction  string          `json:"chapter_instruction"`
	ChapterHeadings     []string        `json:"chapter_headings"`
	FinalInstruction    string          `json:"final_instruction"`
	LocalRunner         string          `json:"local_runner"`
	LocalConfig         json.RawMessage `json:"local_config"`
}

func seedAnalysisCatalog(conn *sql.DB) error {
	if _, err := conn.Exec(`INSERT OR IGNORE INTO app_settings (key, value) VALUES ('analysis_source_rules', ?)`, analysisSourceRules); err != nil {
		return err
	}

	entries, err := loadCatalogEntries()
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := upsertCatalogEntry(conn, e); err != nil {
			return err
		}
	}
	return nil
}

func loadCatalogEntries() ([]catalogEntry, error) {
	names, err := fs.Glob(catalogFS, "catalog/*.json")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no catalog/*.json files embedded")
	}
	out := make([]catalogEntry, 0, len(names))
	for _, name := range names {
		raw, err := catalogFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		var e catalogEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		if e.ID == "" {
			return nil, fmt.Errorf("%s: missing id", name)
		}
		if e.Needs == nil {
			e.Needs = []string{}
		}
		if e.DependsOn == nil {
			e.DependsOn = []string{}
		}
		if e.ChapterHeadings == nil {
			e.ChapterHeadings = []string{}
		}
		if len(e.LocalConfig) == 0 {
			e.LocalConfig = json.RawMessage(`{}`)
		}
		if e.AIProfile == "" {
			e.AIProfile = "single"
		}
		out = append(out, e)
	}
	return out, nil
}

func upsertCatalogEntry(conn *sql.DB, e catalogEntry) error {
	needsJSON, err := json.Marshal(e.Needs)
	if err != nil {
		return fmt.Errorf("%s needs: %w", e.ID, err)
	}
	dependsJSON, err := json.Marshal(e.DependsOn)
	if err != nil {
		return fmt.Errorf("%s depends_on: %w", e.ID, err)
	}
	headingsJSON, err := json.Marshal(e.ChapterHeadings)
	if err != nil {
		return fmt.Errorf("%s chapter_headings: %w", e.ID, err)
	}
	sourceInjection := "all"
	if needsContainStoryElement(e.Needs) {
		sourceInjection = "selective"
	}
	usesAI := 0
	if e.UsesAI {
		usesAI = 1
	}
	usesMerge := 0
	if e.UsesMerge {
		usesMerge = 1
	}
	localRunner := e.LocalRunner
	if e.UsesAI {
		localRunner = ""
	}
	_, err = conn.Exec(`
		INSERT OR REPLACE INTO analysis_catalog (
			id, group_id, group_label, label, description, usage,
			needs, depends_on, uses_ai, uses_merge, ai_profile, sort_order,
			chapter_instruction, chapter_headings, final_instruction,
			source_injection, local_runner, local_config
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.GroupID, e.GroupLabel, e.Label, e.Description, e.Usage,
		string(needsJSON), string(dependsJSON), usesAI, usesMerge, e.AIProfile, e.SortOrder,
		e.ChapterInstruction, string(headingsJSON), e.FinalInstruction,
		sourceInjection, localRunner, string(e.LocalConfig),
	)
	if err != nil {
		return fmt.Errorf("upsert catalog %s: %w", e.ID, err)
	}
	return nil
}

func needsContainStoryElement(needs []string) bool {
	for _, n := range needs {
		switch strings.ToLower(n) {
		case "characters", "locations", "themes":
			return true
		}
	}
	return false
}
