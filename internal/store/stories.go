package store

import (
	"database/sql"
	"fmt"
)

func (s *Store) ListStories(seriesID int64) ([]Story, error) {
	q := `
		SELECT id, series_id, series_sort_order, name, description, COALESCE(pen_name, ''), created_at, updated_at
		FROM stories WHERE user_id = ?`
	args := []any{s.UserID}
	if seriesID > 0 {
		q += ` AND series_id = ?`
		args = append(args, seriesID)
	}
	q += ` ORDER BY series_sort_order IS NULL, series_sort_order, name, updated_at DESC`

	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Story, 0)
	for rows.Next() {
		var item Story
		var sid sql.NullInt64
		var sort sql.NullInt64
		var desc sql.NullString
		if err := rows.Scan(&item.ID, &sid, &sort, &item.Name, &desc, &item.PenName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.SeriesID = int64Ptr(sid)
		item.SeriesSortOrder = intPtr(sort)
		item.Description = strPtr(desc)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) CreateStory(in StoryInput) (IDResult, error) {
	if in.Name == "" {
		return IDResult{}, fmt.Errorf("name is required")
	}
	var series any
	if in.SeriesID > 0 {
		if err := s.ownsSeries(in.SeriesID); err != nil {
			return IDResult{}, fmt.Errorf("series not found")
		}
		series = in.SeriesID
	}
	res, err := s.DB.Exec(`
		INSERT INTO stories (user_id, name, description, series_id, pen_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		s.UserID, in.Name, emptyToNil(in.Description), series, emptyToNil(in.PenName))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	return IDResult{ID: id}, err
}

func (s *Store) UpdateStory(id int64, in StoryInput) (UpdatedResult, error) {
	var series any
	if in.SeriesID > 0 {
		if err := s.ownsSeries(in.SeriesID); err != nil {
			return UpdatedResult{}, fmt.Errorf("series not found")
		}
		series = in.SeriesID
	}
	res, err := s.DB.Exec(`
		UPDATE stories SET
			name = CASE WHEN ? != '' THEN ? ELSE name END,
			description = ?,
			series_id = ?,
			updated_at = datetime('now')
		WHERE id = ? AND user_id = ?`,
		in.Name, in.Name, emptyToNil(in.Description), series, id, s.UserID)
	if err != nil {
		return UpdatedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return UpdatedResult{}, err
	}
	if n == 0 {
		return UpdatedResult{}, fmt.Errorf("story not found")
	}
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) DeleteStory(id int64) (DeletedResult, error) {
	res, err := s.DB.Exec(`DELETE FROM stories WHERE id = ? AND user_id = ?`, id, s.UserID)
	if err != nil {
		return DeletedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return DeletedResult{}, err
	}
	if n == 0 {
		return DeletedResult{}, fmt.Errorf("story not found")
	}
	return DeletedResult{Deleted: true}, nil
}

func (s *Store) ownsStory(storyID int64) error {
	var ok int
	err := s.DB.QueryRow(`SELECT 1 FROM stories WHERE id = ? AND user_id = ?`, storyID, s.UserID).Scan(&ok)
	if err != nil {
		return fmt.Errorf("story not found")
	}
	return nil
}
