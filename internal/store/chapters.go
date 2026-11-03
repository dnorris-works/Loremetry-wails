package store

import (
	"database/sql"
	"fmt"
)

func (s *Store) ListChapters(storyID int64) ([]Chapter, error) {
	if err := s.ownsStory(storyID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`
		SELECT id, story_id, file_name, text_content, sort_order, act_id, created_at
		FROM story_documents
		WHERE story_id = ? AND kind = 'chapter'
		ORDER BY sort_order IS NULL, sort_order, created_at`, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Chapter, 0)
	for rows.Next() {
		ch, err := scanChapter(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func (s *Store) ListStoryDocs(storyID int64) ([]StoryDoc, error) {
	if err := s.ownsStory(storyID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`
		SELECT id, story_id, kind, file_name, sort_order, act_id, created_at
		FROM story_documents
		WHERE story_id = ?
		ORDER BY kind, act_id IS NULL, act_id, sort_order IS NULL, sort_order, created_at`, storyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]StoryDoc, 0)
	for rows.Next() {
		var d StoryDoc
		var sort, act sql.NullInt64
		if err := rows.Scan(&d.ID, &d.StoryID, &d.Kind, &d.FileName, &sort, &act, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.SortOrder = intPtr(sort)
		d.ActID = int64Ptr(act)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) GetChapter(id int64) (Chapter, error) {
	row := s.DB.QueryRow(`
		SELECT sd.id, sd.story_id, sd.file_name, sd.text_content, sd.sort_order, sd.act_id, sd.created_at
		FROM story_documents sd
		JOIN stories st ON st.id = sd.story_id
		WHERE sd.id = ? AND st.user_id = ?`, id, s.UserID)
	ch, err := scanChapter(row)
	if err != nil {
		return Chapter{}, fmt.Errorf("chapter not found")
	}
	return ch, nil
}

func (s *Store) CreateChapter(storyID int64, in ChapterInput) (IDResult, error) {
	if err := s.ownsStory(storyID); err != nil {
		return IDResult{}, err
	}
	if in.FileName == nil || *in.FileName == "" {
		return IDResult{}, fmt.Errorf("file_name is required")
	}
	name := s.chapterCreateName(storyID, *in.FileName)
	res, err := s.DB.Exec(`
		INSERT INTO story_documents (story_id, kind, file_name, mime_type, text_content, sort_order, act_id, created_at)
		VALUES (?, 'chapter', ?, 'text/plain', ?,
			COALESCE(?, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM story_documents WHERE story_id = ? AND kind = 'chapter')),
			?,
			datetime('now'))`,
		storyID, name, nullStr(in.TextContent), nullInt(in.SortOrder), storyID, nullInt64(in.ActID))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	if err == nil {
		s.mirrorWrite("story", storyID, "chapter", "", name, deref(in.TextContent))
	}
	return IDResult{ID: id}, err
}

func (s *Store) CreateStoryDoc(storyID int64, in DocInput) (IDResult, error) {
	if err := s.ownsStory(storyID); err != nil {
		return IDResult{}, err
	}
	if in.Kind == "" || in.FileName == "" {
		return IDResult{}, fmt.Errorf("kind and file_name are required")
	}
	name := in.FileName
	if in.Kind == "chapter" {
		name = s.chapterCreateName(storyID, in.FileName)
	}
	res, err := s.DB.Exec(`
		INSERT INTO story_documents (story_id, kind, file_name, mime_type, text_content, created_at)
		VALUES (?, ?, ?, 'text/plain', ?, datetime('now'))`,
		storyID, in.Kind, name, emptyToNil(in.TextContent))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	if err == nil {
		s.mirrorWrite("story", storyID, in.Kind, "", name, in.TextContent)
	}
	return IDResult{ID: id}, err
}

func (s *Store) UpdateChapter(id int64, in ChapterInput) (UpdatedResult, error) {
	var storyID int64
	var kind, oldName string
	var oldText sql.NullString
	_ = s.DB.QueryRow(`SELECT story_id, kind, file_name, text_content FROM story_documents WHERE id = ?`, id).
		Scan(&storyID, &kind, &oldName, &oldText)
	res, err := s.DB.Exec(`
		UPDATE story_documents SET
			file_name = COALESCE(?, file_name),
			text_content = COALESCE(?, text_content),
			sort_order = COALESCE(?, sort_order),
		    act_id = COALESCE(?, act_id)
		WHERE id = ? AND story_id IN (SELECT id FROM stories WHERE user_id = ?)`,
		nullStr(in.FileName), nullStr(in.TextContent), nullInt(in.SortOrder), nullInt64(in.ActID),
		id, s.UserID)
	if err != nil {
		return UpdatedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return UpdatedResult{}, err
	}
	if n == 0 {
		return UpdatedResult{}, fmt.Errorf("chapter not found")
	}
	name := oldName
	if in.FileName != nil && *in.FileName != "" {
		name = *in.FileName
	}
	text := ""
	if oldText.Valid {
		text = oldText.String
	}
	if in.TextContent != nil {
		text = *in.TextContent
	}
	s.mirrorWrite("story", storyID, kind, oldName, name, text)
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) DeleteChapter(id int64) (DeletedResult, error) {
	var storyID int64
	var kind, name string
	_ = s.DB.QueryRow(`SELECT story_id, kind, file_name FROM story_documents WHERE id = ?`, id).
		Scan(&storyID, &kind, &name)
	res, err := s.DB.Exec(`
		DELETE FROM story_documents
		WHERE id = ? AND story_id IN (SELECT id FROM stories WHERE user_id = ?)`, id, s.UserID)
	if err != nil {
		return DeletedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return DeletedResult{}, err
	}
	if n == 0 {
		return DeletedResult{}, fmt.Errorf("chapter not found")
	}
	s.mirrorRemove("story", storyID, kind, name)
	return DeletedResult{Deleted: true}, nil
}

func (s *Store) chapterCreateName(storyID int64, title string) string {
	if s.folderPath("story", storyID, "chapter") == "" {
		return title
	}
	var max sql.NullInt64
	_ = s.DB.QueryRow(`SELECT MAX(sort_order) FROM story_documents WHERE story_id = ? AND kind = 'chapter'`, storyID).Scan(&max)
	return chapterFileName(int(max.Int64)+1, stripChapterPrefix(title))
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

type scanner interface {
	Scan(dest ...any) error
}

func scanChapter(row scanner) (Chapter, error) {
	var ch Chapter
	var text sql.NullString
	var sort, act sql.NullInt64
	err := row.Scan(&ch.ID, &ch.StoryID, &ch.FileName, &text, &sort, &act, &ch.CreatedAt)
	ch.TextContent = strPtr(text)
	ch.SortOrder = intPtr(sort)
	ch.ActID = int64Ptr(act)
	return ch, err
}
