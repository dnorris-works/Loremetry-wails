package store

import "fmt"

const ReportInlineMax = 50_000

type AnalysisReport struct {
	ID            int64  `json:"id"`
	AnalysisID    string `json:"analysis_id"`
	AnalysisLabel string `json:"analysis_label"`
	ProjectPath   string `json:"project_path"`
	UsesAI        bool   `json:"uses_ai"`
	Body          string `json:"body,omitempty"`
	Large         bool   `json:"large,omitempty"`
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

type ReportBodyRange struct {
	ID     int64  `json:"id"`
	Text   string `json:"text"`
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Total  int    `json:"total"`
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
	var bodySize int
	err := s.DB.QueryRow(`
		SELECT id, analysis_id, analysis_label, project_path, uses_ai, created_at, LENGTH(body)
		FROM analysis_reports WHERE id = ? AND user_id = ?`, id, s.UserID).
		Scan(&r.ID, &r.AnalysisID, &r.AnalysisLabel, &r.ProjectPath, &uses, &r.CreatedAt, &bodySize)
	if err != nil {
		return AnalysisReport{}, fmt.Errorf("report not found")
	}
	r.UsesAI = uses != 0
	r.BodySize = bodySize
	if bodySize > ReportInlineMax {
		r.Large = true
		return r, nil
	}
	err = s.DB.QueryRow(`SELECT body FROM analysis_reports WHERE id = ? AND user_id = ?`, id, s.UserID).
		Scan(&r.Body)
	if err != nil {
		return AnalysisReport{}, fmt.Errorf("report not found")
	}
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

func (s *Store) GetAnalysisReportBodyRange(id int64, offset, limit int) (ReportBodyRange, error) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	var total int
	err := s.DB.QueryRow(`
		SELECT LENGTH(body) FROM analysis_reports WHERE id = ? AND user_id = ?`, id, s.UserID).Scan(&total)
	if err != nil {
		return ReportBodyRange{}, fmt.Errorf("report not found")
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return ReportBodyRange{ID: id, Start: total, End: total, Total: total}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	length := end - offset
	var chunk string
	err = s.DB.QueryRow(`
		SELECT SUBSTR(body, ?, ?) FROM analysis_reports WHERE id = ? AND user_id = ?`,
		offset+1, length, id, s.UserID).Scan(&chunk)
	if err != nil {
		return ReportBodyRange{}, fmt.Errorf("report not found")
	}
	return ReportBodyRange{
		ID:    id,
		Text:  chunk,
		Start: offset,
		End:   end,
		Total: total,
	}, nil
}

func (s *Store) DeleteAnalysisReport(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM analysis_reports WHERE id = ? AND user_id = ?`, id, s.UserID)
	return err
}
