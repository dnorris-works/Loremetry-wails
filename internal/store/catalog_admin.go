package store

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CatalogUpdateInput is the editable subset of an analysis_catalog row.
type CatalogUpdateInput struct {
	ID                 string   `json:"id"`
	Label              string   `json:"label"`
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

func UpdateAnalysisCatalog(in CatalogUpdateInput) (AnalysisDetail, error) {
	if catalogDB == nil {
		return AnalysisDetail{}, fmt.Errorf("catalog database not ready")
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		return AnalysisDetail{}, fmt.Errorf("missing analysis id")
	}
	if _, ok := GetAnalysisDetail(id); !ok {
		return AnalysisDetail{}, fmt.Errorf("unknown analysis")
	}
	needsJSON, err := json.Marshal(nonNilStrings(in.Needs))
	if err != nil {
		return AnalysisDetail{}, fmt.Errorf("needs: %w", err)
	}
	depsJSON, err := json.Marshal(nonNilStrings(in.DependsOn))
	if err != nil {
		return AnalysisDetail{}, fmt.Errorf("depends_on: %w", err)
	}
	headingsJSON, err := json.Marshal(nonNilStrings(in.ChapterHeadings))
	if err != nil {
		return AnalysisDetail{}, fmt.Errorf("chapter_headings: %w", err)
	}
	localConfig := strings.TrimSpace(in.LocalConfig)
	if localConfig == "" {
		localConfig = "{}"
	}
	sourceInjection := strings.TrimSpace(in.SourceInjection)
	if sourceInjection == "" {
		sourceInjection = "selective"
	}
	usesAI := 0
	if in.UsesAI {
		usesAI = 1
	}
	usesMerge := 0
	if in.UsesMerge {
		usesMerge = 1
	}
	_, err = catalogDB.Exec(`
		UPDATE analysis_catalog SET
			label = ?, description = ?, usage = ?,
			needs = ?, depends_on = ?, uses_ai = ?, uses_merge = ?,
			ai_profile = ?, chapter_instruction = ?, chapter_headings = ?,
			final_instruction = ?, source_injection = ?,
			local_runner = ?, local_config = ?
		WHERE id = ?`,
		strings.TrimSpace(in.Label), strings.TrimSpace(in.Description), strings.TrimSpace(in.Usage),
		string(needsJSON), string(depsJSON), usesAI, usesMerge,
		strings.TrimSpace(in.AIProfile), strings.TrimSpace(in.ChapterInstruction), string(headingsJSON),
		strings.TrimSpace(in.FinalInstruction), sourceInjection,
		strings.TrimSpace(in.LocalRunner), localConfig,
		id,
	)
	if err != nil {
		return AnalysisDetail{}, err
	}
	got, ok := GetAnalysisDetail(id)
	if !ok {
		return AnalysisDetail{}, fmt.Errorf("reload catalog row failed")
	}
	return got, nil
}

func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
