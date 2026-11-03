package store

import (
	"fmt"
	"strings"
)

type PlaceInput struct {
	Kind    string  `json:"kind"`
	IDs     []int64 `json:"ids"`
	StoryID int64   `json:"story_id"`
}

type IDList struct {
	IDs []int64 `json:"ids"`
}

func (s *Store) PlaceStoryDoc(id int64, in PlaceInput) (UpdatedResult, error) {
	if in.Kind == "" {
		return UpdatedResult{}, fmt.Errorf("kind is required")
	}
	var storyID int64
	var name string
	err := s.DB.QueryRow(`
		SELECT sd.story_id, sd.file_name
		FROM story_documents sd
		JOIN stories st ON st.id = sd.story_id
		WHERE sd.id = ? AND st.user_id = ?`, id, s.UserID).Scan(&storyID, &name)
	if err != nil {
		return UpdatedResult{}, fmt.Errorf("document not found")
	}
	if in.StoryID > 0 && in.StoryID != storyID {
		if err := s.ownsStory(in.StoryID); err != nil {
			return UpdatedResult{}, err
		}
		storyID = in.StoryID
	}
	name = uniqueFileName(s.storyDocNames(storyID, in.Kind, id), name)
	if _, err := s.DB.Exec(`UPDATE story_documents SET story_id = ?, kind = ?, file_name = ? WHERE id = ?`, storyID, in.Kind, name, id); err != nil {
		return UpdatedResult{}, err
	}
	return s.writeDocOrder(storyID, in.Kind, in.IDs)
}

func (s *Store) PlaceSeriesDoc(id int64, in PlaceInput) (UpdatedResult, error) {
	if in.Kind == "" {
		return UpdatedResult{}, fmt.Errorf("kind is required")
	}
	var seriesID int64
	var name string
	err := s.DB.QueryRow(`
		SELECT d.series_id, d.file_name
		FROM series_bible_documents d
		JOIN series se ON se.id = d.series_id
		WHERE d.id = ? AND se.user_id = ?`, id, s.UserID).Scan(&seriesID, &name)
	if err != nil {
		return UpdatedResult{}, fmt.Errorf("document not found")
	}
	name = uniqueFileName(s.seriesDocNames(seriesID, in.Kind, id), name)
	if _, err := s.DB.Exec(`UPDATE series_bible_documents SET category = ?, file_name = ? WHERE id = ?`, in.Kind, name, id); err != nil {
		return UpdatedResult{}, err
	}
	return s.writeSeriesDocOrder(seriesID, in.Kind, in.IDs)
}

func (s *Store) ReorderActs(storyID int64, in IDList) (UpdatedResult, error) {
	if err := s.ownsStory(storyID); err != nil {
		return UpdatedResult{}, err
	}
	for i, id := range in.IDs {
		if _, err := s.DB.Exec(`
			UPDATE story_acts SET sort_order = ?
			WHERE id = ? AND story_id = ?`, i, id, storyID); err != nil {
			return UpdatedResult{}, err
		}
	}
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) ReorderStories(seriesID int64, in IDList) (UpdatedResult, error) {
	if seriesID > 0 {
		if err := s.ownsSeries(seriesID); err != nil {
			return UpdatedResult{}, err
		}
	}
	for i, id := range in.IDs {
		var err error
		if seriesID > 0 {
			_, err = s.DB.Exec(`
				UPDATE stories SET series_sort_order = ?
				WHERE id = ? AND series_id = ? AND user_id = ?`, i, id, seriesID, s.UserID)
		} else {
			_, err = s.DB.Exec(`
				UPDATE stories SET series_sort_order = ?
				WHERE id = ? AND series_id IS NULL AND user_id = ?`, i, id, s.UserID)
		}
		if err != nil {
			return UpdatedResult{}, err
		}
	}
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) writeDocOrder(storyID int64, kind string, ids []int64) (UpdatedResult, error) {
	for i, id := range ids {
		if _, err := s.DB.Exec(`
			UPDATE story_documents SET sort_order = ?
			WHERE id = ? AND story_id = ? AND kind = ?`, i, id, storyID, kind); err != nil {
			return UpdatedResult{}, err
		}
	}
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) writeSeriesDocOrder(seriesID int64, category string, ids []int64) (UpdatedResult, error) {
	for i, id := range ids {
		if _, err := s.DB.Exec(`
			UPDATE series_bible_documents SET sort_order = ?
			WHERE id = ? AND series_id = ? AND category = ?`, i, id, seriesID, category); err != nil {
			return UpdatedResult{}, err
		}
	}
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) storyDocNames(storyID int64, kind string, exclude int64) []string {
	rows, err := s.DB.Query(`SELECT file_name FROM story_documents WHERE story_id = ? AND kind = ? AND id != ?`, storyID, kind, exclude)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			names = append(names, n)
		}
	}
	if rows.Err() != nil {
		return nil
	}
	return names
}

func (s *Store) seriesDocNames(seriesID int64, category string, exclude int64) []string {
	rows, err := s.DB.Query(`SELECT file_name FROM series_bible_documents WHERE series_id = ? AND category = ? AND id != ?`, seriesID, category, exclude)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			names = append(names, n)
		}
	}
	if rows.Err() != nil {
		return nil
	}
	return names
}

func uniqueFileName(existing []string, name string) string {
	set := map[string]struct{}{}
	for _, n := range existing {
		set[n] = struct{}{}
	}
	if _, ok := set[name]; !ok {
		return name
	}
	base, ext := name, ""
	if i := strings.LastIndex(name, "."); i > 0 {
		base, ext = name[:i], name[i:]
	}
	for n := 2; ; n++ {
		cand := fmt.Sprintf("%s %d%s", base, n, ext)
		if _, ok := set[cand]; !ok {
			return cand
		}
	}
}
