package store

import (
	"fmt"
)

type AnalysisResult struct {
	AnalysisID  string `json:"analysis_id"`
	Label       string `json:"label"`
	ProjectPath string `json:"project_path"`
	DataJSON    string `json:"data_json"`
	Markdown    string `json:"markdown"`
	Original    string `json:"original,omitempty"`
	Proposed    string `json:"proposed,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

// UpsertAnalysisResult stores reusable analysis output for a project (overwrites prior run).
func (s *Store) UpsertAnalysisResult(res AnalysisResult) (AnalysisResult, error) {
	if res.DataJSON == "" {
		res.DataJSON = "{}"
	}
	_, err := s.DB.Exec(`
		INSERT INTO analysis_results (user_id, project_path, analysis_id, label, data_json, markdown, original, proposed, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(user_id, project_path, analysis_id) DO UPDATE SET
			label = excluded.label,
			data_json = excluded.data_json,
			markdown = excluded.markdown,
			original = excluded.original,
			proposed = excluded.proposed,
			updated_at = datetime('now')`,
		s.UserID, res.ProjectPath, res.AnalysisID, res.Label, res.DataJSON, res.Markdown, res.Original, res.Proposed)
	if err != nil {
		return AnalysisResult{}, err
	}
	return s.GetAnalysisResult(res.ProjectPath, res.AnalysisID)
}

func (s *Store) GetAnalysisResult(projectPath, analysisID string) (AnalysisResult, error) {
	var r AnalysisResult
	err := s.DB.QueryRow(`
		SELECT analysis_id, label, project_path, data_json, markdown, COALESCE(original, ''), COALESCE(proposed, ''), updated_at
		FROM analysis_results
		WHERE user_id = ? AND project_path = ? AND analysis_id = ?`,
		s.UserID, projectPath, analysisID).
		Scan(&r.AnalysisID, &r.Label, &r.ProjectPath, &r.DataJSON, &r.Markdown, &r.Original, &r.Proposed, &r.UpdatedAt)
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("analysis result not found")
	}
	return r, nil
}

func (s *Store) ListAnalysisResults(projectPath string) ([]AnalysisResult, error) {
	rows, err := s.DB.Query(`
		SELECT analysis_id, label, project_path, data_json, markdown, COALESCE(original, ''), COALESCE(proposed, ''), updated_at
		FROM analysis_results
		WHERE user_id = ? AND project_path = ?
		ORDER BY updated_at DESC`, s.UserID, projectPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AnalysisResult, 0)
	for rows.Next() {
		var r AnalysisResult
		if err := rows.Scan(&r.AnalysisID, &r.Label, &r.ProjectPath, &r.DataJSON, &r.Markdown, &r.Original, &r.Proposed, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) HasAnalysisResult(projectPath, analysisID string) bool {
	var n int
	err := s.DB.QueryRow(`
		SELECT COUNT(1) FROM analysis_results
		WHERE user_id = ? AND project_path = ? AND analysis_id = ?`,
		s.UserID, projectPath, analysisID).Scan(&n)
	return err == nil && n > 0
}
