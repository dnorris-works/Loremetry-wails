package store

import (
	"database/sql"
	"fmt"
)

func (s *Store) Session() (AuthSession, error) {
	var email, first, last string
	err := s.DB.QueryRow(`SELECT email, first_name, last_name FROM users WHERE id = ?`, s.UserID).
		Scan(&email, &first, &last)
	if err != nil {
		return AuthSession{Authenticated: false, Reason: "local user not found"}, err
	}
	return AuthSession{
		Authenticated: true,
		ID:            fmt.Sprintf("%d", s.UserID),
		Email:         email,
		FirstName:     first,
		LastName:      last,
		IsAdmin:       true,
		BreakGlass:    true,
	}, nil
}

func (s *Store) ListSeries() ([]Series, error) {
	rows, err := s.DB.Query(`
		SELECT id, name, description, summary, COALESCE(pen_name, ''), created_at, updated_at
		FROM series WHERE user_id = ? ORDER BY updated_at DESC`, s.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Series, 0)
	for rows.Next() {
		var item Series
		var desc, summary sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &desc, &summary, &item.PenName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Description = strPtr(desc)
		item.Summary = strPtr(summary)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) CreateSeries(in SeriesInput) (IDResult, error) {
	if in.Name == "" {
		return IDResult{}, fmt.Errorf("name is required")
	}
	res, err := s.DB.Exec(`
		INSERT INTO series (user_id, name, description, summary, pen_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		s.UserID, in.Name, emptyToNil(in.Description), emptyToNil(in.Summary), emptyToNil(in.PenName))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	return IDResult{ID: id}, err
}

func (s *Store) UpdateSeries(id int64, in SeriesInput) (UpdatedResult, error) {
	res, err := s.DB.Exec(`
		UPDATE series SET
			name = CASE WHEN ? != '' THEN ? ELSE name END,
			description = ?,
			summary = ?,
			updated_at = datetime('now')
		WHERE id = ? AND user_id = ?`,
		in.Name, in.Name, emptyToNil(in.Description), emptyToNil(in.Summary), id, s.UserID)
	if err != nil {
		return UpdatedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return UpdatedResult{}, err
	}
	if n == 0 {
		return UpdatedResult{}, fmt.Errorf("series not found")
	}
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) DeleteSeries(id int64) (DeletedResult, error) {
	res, err := s.DB.Exec(`DELETE FROM series WHERE id = ? AND user_id = ?`, id, s.UserID)
	if err != nil {
		return DeletedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return DeletedResult{}, err
	}
	if n == 0 {
		return DeletedResult{}, fmt.Errorf("series not found")
	}
	return DeletedResult{Deleted: true}, nil
}

func (s *Store) GetSeriesBible(seriesID int64) (BibleDoc, error) {
	if err := s.ownsSeries(seriesID); err != nil {
		return BibleDoc{}, err
	}
	var doc BibleDoc
	doc.SeriesID = seriesID
	err := s.DB.QueryRow(`
		SELECT id, COALESCE(text_content, '') FROM series_bible_documents
		WHERE series_id = ? AND category = 'bible' ORDER BY sort_order LIMIT 1`, seriesID).
		Scan(&doc.ID, &doc.TextContent)
	if err == sql.ErrNoRows {
		return BibleDoc{ID: 0, SeriesID: seriesID, TextContent: ""}, nil
	}
	return doc, err
}

func (s *Store) UpdateSeriesBible(seriesID int64, text string) (UpdatedResult, error) {
	if err := s.ownsSeries(seriesID); err != nil {
		return UpdatedResult{}, err
	}
	res, err := s.DB.Exec(`
		UPDATE series_bible_documents SET text_content = ?
		WHERE series_id = ? AND category = 'bible'`, text, seriesID)
	if err != nil {
		return UpdatedResult{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		_, err = s.DB.Exec(`
			INSERT INTO series_bible_documents (series_id, category, file_name, mime_type, text_content, sort_order, created_at)
			VALUES (?, 'bible', 'bible.md', 'text/plain', ?, 0, datetime('now'))`, seriesID, text)
		if err != nil {
			return UpdatedResult{}, err
		}
	}
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) ownsSeries(seriesID int64) error {
	var ok int
	err := s.DB.QueryRow(`SELECT 1 FROM series WHERE id = ? AND user_id = ?`, seriesID, s.UserID).Scan(&ok)
	if err != nil {
		return fmt.Errorf("series not found")
	}
	return nil
}
