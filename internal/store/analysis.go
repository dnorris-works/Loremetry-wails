package store

import (
	"database/sql"
	"encoding/json"
	"strings"
)

var catalogDB *sql.DB

// SetCatalogDB sets the package-level database connection used by catalog queries.
// Must be called once at startup before any catalog access.
func SetCatalogDB(db *sql.DB) {
	catalogDB = db
}

// AIProfile returns the ai_profile value for the given analysis ID.
// Returns "single" if the ID is not found or catalogDB is nil.
func AIProfile(id string) string {
	if catalogDB == nil {
		return "single"
	}
	var profile string
	err := catalogDB.QueryRow(`SELECT ai_profile FROM analysis_catalog WHERE id = ?`, id).Scan(&profile)
	if err != nil {
		return "single"
	}
	return profile
}

type AnalysisItem struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Needs       []string `json:"needs"`
	DependsOn   []string `json:"depends_on"`
	UsesAI      bool     `json:"uses_ai"`
	UsesMerge   bool     `json:"uses_merge"`
}

type AnalysisDetail struct {
	ID                 string   `json:"id"`
	Label              string   `json:"label"`
	Group              string   `json:"group"`
	Description        string   `json:"description"`
	Usage              string   `json:"usage"`
	Needs              []string `json:"needs"`
	DependsOn          []string `json:"depends_on"`
	UsesAI             bool     `json:"uses_ai"`
	UsesMerge          bool     `json:"uses_merge"`
	AIProfile          string   `json:"ai_profile"`
	ChapterInstruction string   `json:"chapter_instruction"`
	ChapterHeadings    []string `json:"chapter_headings"`
	FinalInstruction   string   `json:"final_instruction"`
	SourceInjection    string   `json:"source_injection"`
	LocalRunner        string   `json:"local_runner"`
	LocalConfig        string   `json:"local_config"`
}

type AnalysisGroup struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Items []AnalysisItem `json:"items"`
}

func AnalysisCatalog() []AnalysisGroup {
	if catalogDB == nil {
		return nil
	}
	rows, err := catalogDB.Query(`
		SELECT id, group_id, group_label, label, description, needs, depends_on, uses_ai, uses_merge
		FROM analysis_catalog ORDER BY sort_order`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var groups []AnalysisGroup
	groupIdx := map[string]int{}

	for rows.Next() {
		var id, groupID, groupLabel, label, description, needsJSON, depsJSON string
		var usesAI, usesMerge int
		if err := rows.Scan(&id, &groupID, &groupLabel, &label, &description, &needsJSON, &depsJSON, &usesAI, &usesMerge); err != nil {
			continue
		}

		var needs []string
		if json.Unmarshal([]byte(needsJSON), &needs) != nil || needs == nil {
			needs = []string{}
		}
		var dependsOn []string
		if json.Unmarshal([]byte(depsJSON), &dependsOn) != nil || dependsOn == nil {
			dependsOn = []string{}
		}

		item := AnalysisItem{
			ID:          id,
			Label:       label,
			Description: description,
			Needs:       needs,
			DependsOn:   dependsOn,
			UsesAI:      usesAI != 0,
			UsesMerge:   usesMerge != 0,
		}

		idx, exists := groupIdx[groupID]
		if !exists {
			idx = len(groups)
			groupIdx[groupID] = idx
			groups = append(groups, AnalysisGroup{ID: groupID, Label: groupLabel})
		}
		groups[idx].Items = append(groups[idx].Items, item)
	}

	if rows.Err() != nil {
		return nil
	}

	return groups
}

func GetAnalysisDetail(id string) (AnalysisDetail, bool) {
	if catalogDB == nil {
		return AnalysisDetail{}, false
	}
	id = strings.TrimSpace(id)
	var detail AnalysisDetail
	var needsJSON, depsJSON, headingsJSON string
	var usesAI, usesMerge int
	err := catalogDB.QueryRow(`
		SELECT id, group_label, label, description, usage, needs, depends_on, uses_ai, uses_merge, ai_profile,
			chapter_instruction, chapter_headings, final_instruction,
			source_injection, local_runner, local_config
		FROM analysis_catalog WHERE id = ?`, id).Scan(
		&detail.ID, &detail.Group, &detail.Label, &detail.Description, &detail.Usage,
		&needsJSON, &depsJSON, &usesAI, &usesMerge, &detail.AIProfile,
		&detail.ChapterInstruction, &headingsJSON, &detail.FinalInstruction,
		&detail.SourceInjection, &detail.LocalRunner, &detail.LocalConfig,
	)
	if err != nil {
		return AnalysisDetail{}, false
	}
	if json.Unmarshal([]byte(needsJSON), &detail.Needs) != nil || detail.Needs == nil {
		detail.Needs = []string{}
	}
	if json.Unmarshal([]byte(depsJSON), &detail.DependsOn) != nil || detail.DependsOn == nil {
		detail.DependsOn = []string{}
	}
	if json.Unmarshal([]byte(headingsJSON), &detail.ChapterHeadings) != nil || detail.ChapterHeadings == nil {
		detail.ChapterHeadings = []string{}
	}
	detail.UsesAI = usesAI != 0
	detail.UsesMerge = usesMerge != 0
	return detail, true
}

// CollectPrerequisites returns transitive depends_on ids in dependency order (deps before dependents).
// The primary analysis id itself is not included.
func CollectPrerequisites(id string) []string {
	id = strings.TrimSpace(id)
	visited := map[string]bool{}
	var out []string
	var walk func(string)
	walk = func(cur string) {
		detail, ok := GetAnalysisDetail(cur)
		if !ok {
			return
		}
		for _, dep := range detail.DependsOn {
			dep = strings.TrimSpace(dep)
			if dep == "" || dep == id || visited[dep] {
				continue
			}
			visited[dep] = true
			walk(dep)
			out = append(out, dep)
		}
	}
	walk(id)
	return out
}

// AnalysisRunQueue returns prerequisites + primary in run order.
func AnalysisRunQueue(id string) []string {
	id = strings.TrimSpace(id)
	deps := CollectPrerequisites(id)
	out := make([]string, 0, len(deps)+1)
	out = append(out, deps...)
	out = append(out, id)
	return out
}

// AnalysisSourceRules returns the global fidelity line appended when story-element sources are attached.
func AnalysisSourceRules() string {
	if catalogDB == nil {
		return ""
	}
	var value string
	err := catalogDB.QueryRow(`SELECT value FROM app_settings WHERE key = 'analysis_source_rules'`).Scan(&value)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

// UsesSelectiveSourceInjection reports whether by-chapter AI calls should attach
// only story-element profiles relevant to each chapter.
func UsesSelectiveSourceInjection(d AnalysisDetail) bool {
	switch strings.TrimSpace(d.SourceInjection) {
	case "all":
		return false
	case "selective":
		return true
	default:
		for _, need := range d.Needs {
			switch need {
			case "characters", "locations", "themes":
				return true
			}
		}
		return false
	}
}
