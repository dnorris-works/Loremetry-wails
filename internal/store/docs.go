package store

import (
	"database/sql"
	"fmt"
)

func (s *Store) ListDocumentTypes() ([]DocumentType, error) {
	rows, err := s.DB.Query(`
		SELECT code, display_name, COALESCE(sort_order, 0), applies_to_story, applies_to_series, active, editor, storage
		FROM document_types WHERE active = 1 ORDER BY sort_order, code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]DocumentType, 0)
	for rows.Next() {
		var t DocumentType
		var story, series, active int
		if err := rows.Scan(&t.Code, &t.DisplayName, &t.SortOrder, &story, &series, &active, &t.Editor, &t.Storage); err != nil {
			return nil, err
		}
		t.AppliesToStory = story != 0
		t.AppliesToSeries = series != 0
		t.Active = active != 0
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) ListSeriesDocs(seriesID int64) ([]SeriesDoc, error) {
	if err := s.ownsSeries(seriesID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`
		SELECT id, series_id, category, file_name, created_at
		FROM series_bible_documents WHERE series_id = ?
		ORDER BY category, sort_order IS NULL, sort_order, created_at`, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SeriesDoc, 0)
	for rows.Next() {
		var d SeriesDoc
		if err := rows.Scan(&d.ID, &d.SeriesID, &d.Category, &d.FileName, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) GetSeriesDoc(id int64) (BibleDoc, error) {
	var doc BibleDoc
	err := s.DB.QueryRow(`
		SELECT d.id, d.series_id, d.file_name, COALESCE(d.text_content, '')
		FROM series_bible_documents d
		JOIN series se ON se.id = d.series_id
		WHERE d.id = ? AND se.user_id = ?`, id, s.UserID).
		Scan(&doc.ID, &doc.SeriesID, &doc.FileName, &doc.TextContent)
	if err != nil {
		return BibleDoc{}, fmt.Errorf("document not found")
	}
	return doc, nil
}

func (s *Store) CreateSeriesDoc(seriesID int64, in SeriesDocInput) (IDResult, error) {
	if err := s.ownsSeries(seriesID); err != nil {
		return IDResult{}, err
	}
	if in.Category == "" || in.FileName == "" {
		return IDResult{}, fmt.Errorf("category and file_name are required")
	}
	res, err := s.DB.Exec(`
		INSERT INTO series_bible_documents (series_id, category, file_name, mime_type, text_content, created_at)
		VALUES (?, ?, ?, 'text/plain', ?, datetime('now'))`,
		seriesID, in.Category, in.FileName, emptyToNil(in.TextContent))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	if err == nil {
		s.mirrorWrite("series", seriesID, in.Category, "", in.FileName, in.TextContent)
	}
	return IDResult{ID: id}, err
}

func (s *Store) UpdateSeriesDoc(id int64, in SeriesDocInput) (UpdatedResult, error) {
	var seriesID int64
	var category, oldName string
	_ = s.DB.QueryRow(`SELECT series_id, category, file_name FROM series_bible_documents WHERE id = ?`, id).
		Scan(&seriesID, &category, &oldName)
	res, err := s.DB.Exec(`
		UPDATE series_bible_documents SET
			file_name = CASE WHEN ? != '' THEN ? ELSE file_name END,
			text_content = ?
		WHERE id = ? AND series_id IN (SELECT id FROM series WHERE user_id = ?)`,
		in.FileName, in.FileName, in.TextContent, id, s.UserID)
	if err != nil {
		return UpdatedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return UpdatedResult{}, err
	}
	if n == 0 {
		return UpdatedResult{}, fmt.Errorf("document not found")
	}
	name := in.FileName
	if name == "" {
		name = oldName
	}
	s.mirrorWrite("series", seriesID, category, oldName, name, in.TextContent)
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) DeleteSeriesDoc(id int64) (DeletedResult, error) {
	var seriesID int64
	var category, name string
	_ = s.DB.QueryRow(`SELECT series_id, category, file_name FROM series_bible_documents WHERE id = ?`, id).
		Scan(&seriesID, &category, &name)
	res, err := s.DB.Exec(`
		DELETE FROM series_bible_documents
		WHERE id = ? AND series_id IN (SELECT id FROM series WHERE user_id = ?)`, id, s.UserID)
	if err != nil {
		return DeletedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return DeletedResult{}, err
	}
	if n == 0 {
		return DeletedResult{}, fmt.Errorf("document not found")
	}
	s.mirrorRemove("series", seriesID, category, name)
	return DeletedResult{Deleted: true}, nil
}

func (s *Store) ListActs(storyID int64) ([]StoryAct, error) {
	if err := s.ownsStory(storyID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`
		SELECT id, story_id, title, summary, sort_order, created_at, updated_at
		FROM story_acts WHERE story_id = ?
		ORDER BY sort_order IS NULL, sort_order, created_at`, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]StoryAct, 0)
	for rows.Next() {
		act, err := scanAct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, act)
	}
	return out, rows.Err()
}

func (s *Store) GetAct(id int64) (StoryAct, error) {
	row := s.DB.QueryRow(`
		SELECT a.id, a.story_id, a.title, a.summary, a.sort_order, a.created_at, a.updated_at
		FROM story_acts a
		JOIN stories st ON st.id = a.story_id
		WHERE a.id = ? AND st.user_id = ?`, id, s.UserID)
	act, err := scanAct(row)
	if err != nil {
		return StoryAct{}, fmt.Errorf("act not found")
	}
	return act, nil
}

func (s *Store) CreateAct(storyID int64, in ActInput) (IDResult, error) {
	if err := s.ownsStory(storyID); err != nil {
		return IDResult{}, err
	}
	if in.Title == "" {
		return IDResult{}, fmt.Errorf("title is required")
	}
	res, err := s.DB.Exec(`
		INSERT INTO story_acts (story_id, title, summary, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM story_acts WHERE story_id = ?), datetime('now'), datetime('now'))`,
		storyID, in.Title, emptyToNil(in.Summary), storyID)
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	if err == nil {
		s.mirrorWrite("story", storyID, "act", "", in.Title, in.Summary)
	}
	return IDResult{ID: id}, err
}

func (s *Store) UpdateAct(id int64, in ActInput) (UpdatedResult, error) {
	var storyID int64
	var oldTitle string
	_ = s.DB.QueryRow(`SELECT story_id, title FROM story_acts WHERE id = ?`, id).Scan(&storyID, &oldTitle)
	res, err := s.DB.Exec(`
		UPDATE story_acts SET
			title = CASE WHEN ? != '' THEN ? ELSE title END,
			summary = COALESCE(?, summary),
			updated_at = datetime('now')
		WHERE id = ? AND story_id IN (SELECT id FROM stories WHERE user_id = ?)`,
		in.Title, in.Title, emptyToNil(in.Summary), id, s.UserID)
	if err != nil {
		return UpdatedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return UpdatedResult{}, err
	}
	if n == 0 {
		return UpdatedResult{}, fmt.Errorf("act not found")
	}
	title := in.Title
	if title == "" {
		title = oldTitle
	}
	s.mirrorWrite("story", storyID, "act", oldTitle, title, in.Summary)
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) DeleteAct(id int64) (DeletedResult, error) {
	var storyID int64
	var title string
	_ = s.DB.QueryRow(`SELECT story_id, title FROM story_acts WHERE id = ?`, id).Scan(&storyID, &title)
	res, err := s.DB.Exec(`
		DELETE FROM story_acts
		WHERE id = ? AND story_id IN (SELECT id FROM stories WHERE user_id = ?)`, id, s.UserID)
	if err != nil {
		return DeletedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return DeletedResult{}, err
	}
	if n == 0 {
		return DeletedResult{}, fmt.Errorf("act not found")
	}
	s.mirrorRemove("story", storyID, "act", title)
	return DeletedResult{Deleted: true}, nil
}

func scanAct(row scanner) (StoryAct, error) {
	var a StoryAct
	var summary sql.NullString
	var sort sql.NullInt64
	err := row.Scan(&a.ID, &a.StoryID, &a.Title, &summary, &sort, &a.CreatedAt, &a.UpdatedAt)
	a.Summary = strPtr(summary)
	a.SortOrder = intPtr(sort)
	return a, err
}
