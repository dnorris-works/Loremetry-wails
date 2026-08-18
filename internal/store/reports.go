package store

import "fmt"

type AnalysisReport struct {
	ID            int64  `json:"id"`
	AnalysisID    string `json:"analysis_id"`
	AnalysisLabel string `json:"analysis_label"`
	ProjectPath   string `json:"project_path"`
	UsesAI        bool   `json:"uses_ai"`
	Body          string `json:"body"`
	CreatedAt     string `json:"created_at"`
}

func (s *Store) SaveAnalysisReport(rep AnalysisReport) (AnalysisReport, error) {
	uses := 0
	if rep.UsesAI {
		uses = 1
	}
	res, err := s.DB.Exec(`
		INSERT INTO analysis_reports (user_id, analysis_id, analysis_label, project_path, uses_ai, body, created_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
		s.UserID, rep.AnalysisID, rep.AnalysisLabel, rep.ProjectPath, uses, rep.Body)
	if err != nil {
		return AnalysisReport{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return AnalysisReport{}, err
	}
	return s.GetAnalysisReport(id)
}

func (s *Store) GetAnalysisReport(id int64) (AnalysisReport, error) {
	var r AnalysisReport
	var uses int
	err := s.DB.QueryRow(`
		SELECT id, analysis_id, analysis_label, project_path, uses_ai, body, created_at
		FROM analysis_reports WHERE id = ? AND user_id = ?`, id, s.UserID).
		Scan(&r.ID, &r.AnalysisID, &r.AnalysisLabel, &r.ProjectPath, &uses, &r.Body, &r.CreatedAt)
	if err != nil {
		return AnalysisReport{}, fmt.Errorf("report not found")
	}
	r.UsesAI = uses != 0
	return r, nil
}

func (s *Store) ListAnalysisReports() ([]AnalysisReport, error) {
	rows, err := s.DB.Query(`
		SELECT id, analysis_id, analysis_label, project_path, uses_ai, body, created_at
		FROM analysis_reports WHERE user_id = ? ORDER BY created_at DESC, id DESC`, s.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AnalysisReport, 0)
	for rows.Next() {
		var r AnalysisReport
		var uses int
		if err := rows.Scan(&r.ID, &r.AnalysisID, &r.AnalysisLabel, &r.ProjectPath, &uses, &r.Body, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.UsesAI = uses != 0
		out = append(out, r)
	}
	return out, rows.Err()
}
