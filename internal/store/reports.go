package store

import (
	"fmt"
	"strings"
)

// FormatReportBody prepends the standard About / How to use header to report content.
// The full markdown (header + body) is what gets saved in analysis_reports.body.
func FormatReportBody(label, analysisID, content string) string {
	about := strings.TrimSpace(analysisDescription(analysisID))
	usage := strings.TrimSpace(analysisUsage(analysisID))
	content = strings.TrimSpace(content)
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", strings.TrimSpace(label))
	b.WriteString("## About this report\n\n")
	if about != "" {
		b.WriteString(about)
		b.WriteString("\n\n")
	}
	b.WriteString("## How to use\n\n")
	if usage != "" {
		b.WriteString(usage)
		b.WriteString("\n\n")
	}
	b.WriteString("---\n\n")
	if content != "" {
		b.WriteString(content)
		b.WriteByte('\n')
	}
	return b.String()
}

type AnalysisReport struct {
	ID            int64  `json:"id"`
	AnalysisID    string `json:"analysis_id"`
	AnalysisLabel string `json:"analysis_label"`
	ProjectPath   string `json:"project_path"`
	UsesAI        bool   `json:"uses_ai"`
	Body          string `json:"body,omitempty"`
	Original      string `json:"original,omitempty"`
	Proposed      string `json:"proposed,omitempty"`
	BodySize      int    `json:"body_size,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type AnalysisReportSummary struct {
	ID            int64  `json:"id"`
	AnalysisID    string `json:"analysis_id"`
	AnalysisLabel string `json:"analysis_label"`
	ProjectPath   string `json:"project_path"`
	UsesAI        bool   `json:"uses_ai"`
	BodySize      int    `json:"body_size"`
	CreatedAt     string `json:"created_at"`
}

func (s *Store) SaveAnalysisReport(rep AnalysisReport) (AnalysisReport, error) {
	uses := 0
	if rep.UsesAI {
		uses = 1
	}
	res, err := s.DB.Exec(`
		INSERT INTO analysis_reports (user_id, analysis_id, analysis_label, project_path, uses_ai, body, original, proposed, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		s.UserID, rep.AnalysisID, rep.AnalysisLabel, rep.ProjectPath, uses, rep.Body, rep.Original, rep.Proposed)
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
		SELECT id, analysis_id, analysis_label, project_path, uses_ai, body, COALESCE(original, ''), COALESCE(proposed, ''), created_at
		FROM analysis_reports WHERE id = ? AND user_id = ?`, id, s.UserID).
		Scan(&r.ID, &r.AnalysisID, &r.AnalysisLabel, &r.ProjectPath, &uses, &r.Body, &r.Original, &r.Proposed, &r.CreatedAt)
	if err != nil {
		return AnalysisReport{}, fmt.Errorf("report not found")
	}
	r.UsesAI = uses != 0
	r.BodySize = len(r.Body)
	return r, nil
}

func (s *Store) ListAnalysisReports() ([]AnalysisReportSummary, error) {
	rows, err := s.DB.Query(`
		SELECT id, analysis_id, analysis_label, project_path, uses_ai, created_at, LENGTH(body)
		FROM analysis_reports WHERE user_id = ? ORDER BY created_at DESC, id DESC`, s.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AnalysisReportSummary, 0)
	for rows.Next() {
		var r AnalysisReportSummary
		var uses int
		if err := rows.Scan(&r.ID, &r.AnalysisID, &r.AnalysisLabel, &r.ProjectPath, &uses, &r.CreatedAt, &r.BodySize); err != nil {
			return nil, err
		}
		r.UsesAI = uses != 0
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) DeleteAnalysisReport(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM analysis_reports WHERE id = ? AND user_id = ?`, id, s.UserID)
	return err
}
