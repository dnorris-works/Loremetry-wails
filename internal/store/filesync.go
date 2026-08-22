package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var chapterNumRE = regexp.MustCompile(`^(\d+)[-_.\s]+(.+)$`)
var unsafeFileRE = regexp.MustCompile(`[\\/:*?"<>|]`)

type diskFile struct {
	Name string
	Rel  string
	Text string
}

func (s *Store) SyncAllFolders() ([]FolderChange, error) {
	var changes []FolderChange
	links, err := s.AllFolderLinks()
	if err != nil {
		return nil, err
	}
	for _, l := range links {
		part, err := s.SyncFolder(l.Scope, l.OwnerID, l.Kind)
		if err != nil {
			return nil, err
		}
		changes = append(changes, part...)
	}
	if changes == nil {
		changes = []FolderChange{}
	}
	return changes, nil
}

func (s *Store) SyncFolder(scope string, ownerID int64, kind string) ([]FolderChange, error) {
	path := s.folderPath(scope, ownerID, kind)
	if path == "" || kind == "root" {
		return nil, nil
	}
	files, err := listTextFiles(path)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "character":
		return s.syncCharacters(scope, ownerID, path, files)
	case "act":
		if scope != "story" {
			return nil, nil
		}
		return s.syncActs(ownerID, path, files)
	case "chapter":
		if scope != "story" {
			return nil, nil
		}
		return s.syncChapters(ownerID, path, files)
	default:
		return s.syncDocs(scope, ownerID, kind, path, files)
	}
}

func (s *Store) exportIfFolderEmpty(scope string, ownerID int64, kind, path string) error {
	files, err := listTextFiles(path)
	if err != nil {
		return err
	}
	if len(files) > 0 {
		return nil
	}
	return s.exportKind(scope, ownerID, kind, path)
}

func (s *Store) exportKind(scope string, ownerID int64, kind, path string) error {
	switch kind {
	case "character":
		var rows []CharacterProfile
		var err error
		if scope == "series" {
			rows, err = s.ListSeriesCharacters(ownerID)
		} else {
			rows, err = s.ListCharacters(ownerID)
		}
		if err != nil {
			return err
		}
		for _, c := range rows {
			name := mdName(c.CharacterName)
			text := characterFileText(c.CharacterName, strOr(c.CharacterVoiceNotes))
			if err := writeTextFile(filepath.Join(path, name), text); err != nil {
				return err
			}
			_, _ = s.DB.Exec(`UPDATE character_profiles SET source_rel = ? WHERE id = ?`, name, c.ID)
		}
	case "act":
		rows, err := s.ListActs(ownerID)
		if err != nil {
			return err
		}
		for i, a := range rows {
			name := chapterFileName(i+1, a.Title)
			if err := writeTextFile(filepath.Join(path, name), strOr(a.Summary)); err != nil {
				return err
			}
			_, _ = s.DB.Exec(`UPDATE story_acts SET source_rel = ? WHERE id = ?`, name, a.ID)
		}
	case "chapter":
		rows, err := s.ListChapters(ownerID)
		if err != nil {
			return err
		}
		for i, ch := range rows {
			title := stripChapterPrefix(ch.FileName)
			name := chapterFileName(i+1, title)
			if err := writeTextFile(filepath.Join(path, name), strOr(ch.TextContent)); err != nil {
				return err
			}
			_, _ = s.DB.Exec(`UPDATE story_documents SET file_name = ?, sort_order = ?, source_rel = ? WHERE id = ?`, name, i+1, name, ch.ID)
		}
	default:
		if scope == "series" {
			docs, err := s.ListSeriesDocs(ownerID)
			if err != nil {
				return err
			}
			for _, d := range docs {
				if d.Category != kind {
					continue
				}
				doc, err := s.GetSeriesDoc(d.ID)
				if err != nil {
					return err
				}
				name := mdName(d.FileName)
				if err := writeTextFile(filepath.Join(path, name), doc.TextContent); err != nil {
					return err
				}
				_, _ = s.DB.Exec(`UPDATE series_bible_documents SET source_rel = ? WHERE id = ?`, name, d.ID)
			}
			return nil
		}
		docs, err := s.ListStoryDocs(ownerID)
		if err != nil {
			return err
		}
		for _, d := range docs {
			if d.Kind != kind {
				continue
			}
			ch, err := s.GetChapter(d.ID)
			if err != nil {
				return err
			}
			name := mdName(d.FileName)
			if err := writeTextFile(filepath.Join(path, name), strOr(ch.TextContent)); err != nil {
				return err
			}
			_, _ = s.DB.Exec(`UPDATE story_documents SET source_rel = ? WHERE id = ?`, name, d.ID)
		}
	}
	return nil
}

func (s *Store) syncDocs(scope string, ownerID int64, kind, dir string, files []diskFile) ([]FolderChange, error) {
	var rows []namedRow
	var changes []FolderChange
	if scope == "series" {
		docs, err := s.ListSeriesDocs(ownerID)
		if err != nil {
			return nil, err
		}
		rels := s.docSourceRels("series_bible_documents", "series_id", ownerID, "category", kind)
		for _, d := range docs {
			if d.Category == kind {
				rows = append(rows, namedRow{d.ID, d.FileName, rels[d.ID]})
			}
		}
	} else {
		docs, err := s.ListStoryDocs(ownerID)
		if err != nil {
			return nil, err
		}
		rels := s.docSourceRels("story_documents", "story_id", ownerID, "kind", kind)
		for _, d := range docs {
			if d.Kind == kind {
				rows = append(rows, namedRow{d.ID, d.FileName, rels[d.ID]})
			}
		}
	}
	typ := "doc"
	if scope == "series" {
		typ = "series-doc"
	}
	seen := map[int64]bool{}
	used := map[string]bool{}
	for _, f := range files {
		id := matchFile(rows, f, seen)
		name := uniqueDisplayName(used, f.Rel, f.Name)
		if id == 0 {
			if scope == "series" {
				if _, err := s.insertSeriesDocOnly(ownerID, kind, name, f.Text, f.Rel); err != nil {
					return nil, err
				}
			} else if _, err := s.insertStoryDocOnly(ownerID, kind, name, f.Text, f.Rel); err != nil {
				return nil, err
			}
			continue
		}
		oldName, oldText := s.docNameText(scope, id)
		if err := s.updateDocText(scope, id, name, f.Text, f.Rel); err != nil {
			return nil, err
		}
		if oldName != name || oldText != f.Text {
			changes = append(changes, FolderChange{Type: typ, ID: id})
		}
	}
	for _, r := range rows {
		if seen[r.id] {
			continue
		}
		if scope == "series" {
			if _, err := s.deleteSeriesDocOnly(r.id); err != nil {
				return nil, err
			}
		} else if _, err := s.deleteStoryDocOnly(r.id); err != nil {
			return nil, err
		}
		changes = append(changes, FolderChange{Type: typ, ID: r.id})
	}
	_ = dir
	return changes, nil
}

func (s *Store) syncCharacters(scope string, ownerID int64, dir string, files []diskFile) ([]FolderChange, error) {
	var list []CharacterProfile
	var err error
	if scope == "series" {
		list, err = s.ListSeriesCharacters(ownerID)
	} else {
		list, err = s.ListCharacters(ownerID)
	}
	if err != nil {
		return nil, err
	}
	rels := s.charSourceRels(scope, ownerID)
	rows := make([]namedRow, 0, len(list))
	for _, c := range list {
		rows = append(rows, namedRow{c.ID, c.CharacterName, rels[c.ID]})
	}
	var changes []FolderChange
	seen := map[int64]bool{}
	used := map[string]bool{}
	for _, f := range files {
		name, notes := parseNamedFile(f.Name, f.Text)
		id := matchFile(rows, f, seen)
		if id == 0 {
			id = matchRow(rows, name, seen)
		}
		display := uniqueDisplayName(used, f.Rel, name)
		importance := characterImportanceFromRel(f.Rel)
		if id == 0 {
			in := CharacterInput{CharacterName: display, CharacterVoiceNotes: notes, StoryImportance: importance}
			if scope == "series" {
				in.SeriesID = ownerID
			} else {
				in.StoryID = ownerID
			}
			if _, err := s.insertCharacterOnly(in, f.Rel); err != nil {
				return nil, err
			}
			continue
		}
		old, err := s.GetCharacter(id)
		if err != nil {
			return nil, err
		}
		if _, err := s.DB.Exec(`
			UPDATE character_profiles SET character_name = ?, character_voice_notes = ?, story_importance = ?, source_rel = ?, updated_at = datetime('now')
			WHERE id = ?`, display, emptyToNil(notes), emptyToNil(importance), f.Rel, id); err != nil {
			return nil, err
		}
		if old.CharacterName != display || strOr(old.CharacterVoiceNotes) != notes || strOr(old.StoryImportance) != importance {
			changes = append(changes, FolderChange{Type: "character", ID: id})
		}
	}
	for _, r := range rows {
		if !seen[r.id] {
			if _, err := s.deleteCharacterOnly(r.id); err != nil {
				return nil, err
			}
			changes = append(changes, FolderChange{Type: "character", ID: r.id})
		}
	}
	_ = dir
	return changes, nil
}

func (s *Store) syncActs(storyID int64, dir string, files []diskFile) ([]FolderChange, error) {
	list, err := s.ListActs(storyID)
	if err != nil {
		return nil, err
	}
	rels := s.actSourceRels(storyID)
	rows := make([]namedRow, 0, len(list))
	for _, a := range list {
		rows = append(rows, namedRow{a.ID, a.Title, rels[a.ID]})
	}
	sort.Slice(files, func(i, j int) bool {
		return chapterFileSort(files[i]) < chapterFileSort(files[j])
	})
	var changes []FolderChange
	seen := map[int64]bool{}
	used := map[string]bool{}
	for i, f := range files {
		title, summary := parseNamedFile(f.Name, f.Text)
		title = stripChapterPrefix(title)
		id := matchFile(rows, f, seen)
		if id == 0 {
			id = matchRow(rows, title, seen)
		}
		display := uniqueDisplayName(used, f.Rel, title)
		wantRel := relWithBase(f.Rel, chapterFileName(i+1, title))
		if id == 0 {
			if _, err := s.insertActOnly(storyID, display, summary, i+1, wantRel); err != nil {
				return nil, err
			}
			continue
		}
		old, err := s.GetAct(id)
		if err != nil {
			return nil, err
		}
		if _, err := s.DB.Exec(`
			UPDATE story_acts SET title = ?, summary = ?, sort_order = ?, source_rel = ?, updated_at = datetime('now')
			WHERE id = ?`, display, emptyToNil(summary), i+1, wantRel, id); err != nil {
			return nil, err
		}
		if old.Title != display || strOr(old.Summary) != summary {
			changes = append(changes, FolderChange{Type: "act", ID: id})
		}
	}
	for _, r := range rows {
		if !seen[r.id] {
			if _, err := s.deleteActOnly(r.id); err != nil {
				return nil, err
			}
			changes = append(changes, FolderChange{Type: "act", ID: r.id})
		}
	}
	if err := s.rewriteNumberedFiles(dir, files, func(i int, f diskFile) string {
		title, _ := parseNamedFile(f.Name, f.Text)
		return chapterFileName(i+1, stripChapterPrefix(title))
	}); err != nil {
		return nil, err
	}
	return changes, nil
}

func (s *Store) syncChapters(storyID int64, dir string, files []diskFile) ([]FolderChange, error) {
	list, err := s.ListChapters(storyID)
	if err != nil {
		return nil, err
	}
	rels := s.docSourceRels("story_documents", "story_id", storyID, "kind", "chapter")
	rows := make([]namedRow, 0, len(list))
	for _, ch := range list {
		rows = append(rows, namedRow{ch.ID, ch.FileName, rels[ch.ID]})
	}
	sort.Slice(files, func(i, j int) bool {
		return chapterFileSort(files[i]) < chapterFileSort(files[j])
	})
	var changes []FolderChange
	seen := map[int64]bool{}
	used := map[string]bool{}
	for i, f := range files {
		title := stripChapterPrefix(f.Name)
		id := matchFile(rows, f, seen)
		if id == 0 {
			id = matchRow(rows, title, seen)
		}
		name := uniqueDisplayName(used, f.Rel, chapterFileName(i+1, title))
		wantRel := relWithBase(f.Rel, chapterFileName(i+1, title))
		if id == 0 {
			res, err := s.insertStoryDocOnly(storyID, "chapter", name, f.Text, wantRel)
			if err != nil {
				return nil, err
			}
			id = res.ID
		} else {
			old, err := s.GetChapter(id)
			if err != nil {
				return nil, err
			}
			if old.FileName != name || strOr(old.TextContent) != f.Text {
				changes = append(changes, FolderChange{Type: "doc", ID: id})
			}
		}
		if _, err := s.DB.Exec(`
			UPDATE story_documents SET file_name = ?, text_content = ?, sort_order = ?, source_rel = ?
			WHERE id = ?`, name, f.Text, i+1, wantRel, id); err != nil {
			return nil, err
		}
	}
	for _, r := range rows {
		if !seen[r.id] {
			if _, err := s.deleteStoryDocOnly(r.id); err != nil {
				return nil, err
			}
			changes = append(changes, FolderChange{Type: "doc", ID: r.id})
		}
	}
	if err := s.rewriteNumberedFiles(dir, files, func(i int, f diskFile) string {
		return chapterFileName(i+1, stripChapterPrefix(f.Name))
	}); err != nil {
		return nil, err
	}
	return changes, nil
}

func (s *Store) docNameText(scope string, id int64) (string, string) {
	if scope == "series" {
		doc, err := s.GetSeriesDoc(id)
		if err != nil {
			return "", ""
		}
		return doc.FileName, doc.TextContent
	}
	ch, err := s.GetChapter(id)
	if err != nil {
		return "", ""
	}
	return ch.FileName, strOr(ch.TextContent)
}

func (s *Store) rewriteNumberedFiles(dir string, files []diskFile, nameAt func(int, diskFile) string) error {
	type step struct{ from, to string }
	var temps []step
	for i, f := range files {
		want := relWithBase(f.Rel, nameAt(i, f))
		if slashRel(f.Rel) == want {
			continue
		}
		tmp := relWithBase(f.Rel, fmt.Sprintf(".renumber-%d-%s", i, filepath.Base(want)))
		if err := os.Rename(filepath.Join(dir, filepath.FromSlash(f.Rel)), filepath.Join(dir, filepath.FromSlash(tmp))); err != nil {
			return err
		}
		temps = append(temps, step{tmp, want})
	}
	for _, t := range temps {
		if err := os.Rename(filepath.Join(dir, filepath.FromSlash(t.from)), filepath.Join(dir, filepath.FromSlash(t.to))); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) mirrorWrite(scope string, ownerID int64, kind, oldName, newName, text string) {
	dir := s.folderPath(scope, ownerID, kind)
	if dir == "" {
		return
	}
	name := newName
	switch kind {
	case "character":
		text = characterFileText(newName, text)
		name = mdName(newName)
	case "chapter", "act":
		if n, title, ok := splitChapterName(newName); ok {
			name = chapterFileName(n, title)
		} else {
			name = mdName(newName)
		}
	default:
		name = mdName(newName)
	}
	rel := s.lookupSourceRel(scope, ownerID, kind, oldName)
	if rel == "" {
		rel = s.lookupSourceRel(scope, ownerID, kind, newName)
	}
	destRel := name
	if rel != "" {
		destRel = relWithBase(rel, name)
	}
	_ = writeTextFile(filepath.Join(dir, filepath.FromSlash(destRel)), text)
	if rel != "" && slashRel(rel) != destRel {
		_ = os.Remove(filepath.Join(dir, filepath.FromSlash(rel)))
	} else if oldName != "" && matchKey(oldName) != matchKey(name) {
		_ = os.Remove(filepath.Join(dir, oldName))
		_ = os.Remove(filepath.Join(dir, mdName(oldName)))
	}
	s.setSourceRelByName(scope, ownerID, kind, newName, destRel)
}

func (s *Store) mirrorRemove(scope string, ownerID int64, kind, name string) {
	dir := s.folderPath(scope, ownerID, kind)
	if dir == "" || name == "" {
		return
	}
	rel := s.lookupSourceRel(scope, ownerID, kind, name)
	if rel != "" {
		_ = os.Remove(filepath.Join(dir, filepath.FromSlash(rel)))
	} else {
		_ = os.Remove(filepath.Join(dir, name))
		_ = os.Remove(filepath.Join(dir, mdName(name)))
	}
	if kind == "chapter" {
		files, err := listTextFiles(dir)
		if err == nil {
			_, _ = s.syncChapters(ownerID, dir, files)
		}
	}
}

func (s *Store) insertSeriesDocOnly(seriesID int64, category, name, text, rel string) (IDResult, error) {
	res, err := s.DB.Exec(`
		INSERT INTO series_bible_documents (series_id, category, file_name, mime_type, text_content, source_rel, created_at)
		VALUES (?, ?, ?, 'text/plain', ?, ?, datetime('now'))`,
		seriesID, category, name, emptyToNil(text), emptyToNil(rel))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	return IDResult{ID: id}, err
}

func (s *Store) insertStoryDocOnly(storyID int64, kind, name, text, rel string) (IDResult, error) {
	res, err := s.DB.Exec(`
		INSERT INTO story_documents (story_id, kind, file_name, mime_type, text_content, sort_order, source_rel, created_at)
		VALUES (?, ?, ?, 'text/plain', ?,
			(SELECT COALESCE(MAX(sort_order), 0) + 1 FROM story_documents WHERE story_id = ? AND kind = ?),
			?, datetime('now'))`,
		storyID, kind, name, emptyToNil(text), storyID, kind, emptyToNil(rel))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	return IDResult{ID: id}, err
}

func (s *Store) updateDocText(scope string, id int64, name, text, rel string) error {
	if scope == "series" {
		_, err := s.DB.Exec(`UPDATE series_bible_documents SET file_name = ?, text_content = ?, source_rel = ? WHERE id = ?`, name, text, emptyToNil(rel), id)
		return err
	}
	_, err := s.DB.Exec(`UPDATE story_documents SET file_name = ?, text_content = ?, source_rel = ? WHERE id = ?`, name, text, emptyToNil(rel), id)
	return err
}

func (s *Store) deleteSeriesDocOnly(id int64) (DeletedResult, error) {
	_, err := s.DB.Exec(`DELETE FROM series_bible_documents WHERE id = ?`, id)
	return DeletedResult{Deleted: true}, err
}

func (s *Store) deleteStoryDocOnly(id int64) (DeletedResult, error) {
	_, err := s.DB.Exec(`DELETE FROM story_documents WHERE id = ?`, id)
	return DeletedResult{Deleted: true}, err
}

func (s *Store) insertCharacterOnly(in CharacterInput, rel string) (IDResult, error) {
	res, err := s.DB.Exec(`
		INSERT INTO character_profiles (series_id, story_id, character_name, story_importance, character_voice_notes, pov_character, source_rel, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?, datetime('now'), datetime('now'))`,
		zeroToNilInt64(in.SeriesID), zeroToNilInt64(in.StoryID), in.CharacterName, emptyToNil(in.StoryImportance), emptyToNil(in.CharacterVoiceNotes), emptyToNil(rel))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	return IDResult{ID: id}, err
}

func (s *Store) deleteCharacterOnly(id int64) (DeletedResult, error) {
	_, err := s.DB.Exec(`DELETE FROM character_profiles WHERE id = ?`, id)
	return DeletedResult{Deleted: true}, err
}

func (s *Store) insertActOnly(storyID int64, title, summary string, order int, rel string) (IDResult, error) {
	res, err := s.DB.Exec(`
		INSERT INTO story_acts (story_id, title, summary, sort_order, source_rel, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		storyID, title, emptyToNil(summary), order, emptyToNil(rel))
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	return IDResult{ID: id}, err
}

func (s *Store) deleteActOnly(id int64) (DeletedResult, error) {
	_, err := s.DB.Exec(`DELETE FROM story_acts WHERE id = ?`, id)
	return DeletedResult{Deleted: true}, err
}

func (s *Store) docSourceRels(table, ownerCol string, ownerID int64, kindCol, kind string) map[int64]string {
	out := map[int64]string{}
	var q string
	switch table + "|" + ownerCol + "|" + kindCol {
	case "series_bible_documents|series_id|category":
		q = `SELECT id, COALESCE(source_rel, '') FROM series_bible_documents WHERE series_id = ? AND category = ?`
	case "story_documents|story_id|kind":
		q = `SELECT id, COALESCE(source_rel, '') FROM story_documents WHERE story_id = ? AND kind = ?`
	default:
		return out
	}
	rows, err := s.DB.Query(q, ownerID, kind)
	if err != nil {
		return out
	}
	defer rows.Close()
	return scanIDRels(rows, out)
}

func (s *Store) charSourceRels(scope string, ownerID int64) map[int64]string {
	out := map[int64]string{}
	q := `SELECT id, COALESCE(source_rel, '') FROM character_profiles WHERE series_id = ?`
	if scope == "story" {
		q = `SELECT id, COALESCE(source_rel, '') FROM character_profiles WHERE story_id = ?`
	}
	rows, err := s.DB.Query(q, ownerID)
	if err != nil {
		return out
	}
	defer rows.Close()
	return scanIDRels(rows, out)
}

func (s *Store) actSourceRels(storyID int64) map[int64]string {
	out := map[int64]string{}
	rows, err := s.DB.Query(`SELECT id, COALESCE(source_rel, '') FROM story_acts WHERE story_id = ?`, storyID)
	if err != nil {
		return out
	}
	defer rows.Close()
	return scanIDRels(rows, out)
}

func scanIDRels(rows *sql.Rows, out map[int64]string) map[int64]string {
	for rows.Next() {
		var id int64
		var rel string
		if rows.Scan(&id, &rel) == nil {
			out[id] = rel
		}
	}
	_ = rows.Err()
	return out
}

func (s *Store) lookupSourceRel(scope string, ownerID int64, kind, name string) string {
	var rel string
	switch kind {
	case "character":
		q := `SELECT COALESCE(source_rel, '') FROM character_profiles WHERE series_id = ? AND character_name = ?`
		if scope == "story" {
			q = `SELECT COALESCE(source_rel, '') FROM character_profiles WHERE story_id = ? AND character_name = ?`
		}
		_ = s.DB.QueryRow(q, ownerID, name).Scan(&rel)
	case "act":
		_ = s.DB.QueryRow(`SELECT COALESCE(source_rel, '') FROM story_acts WHERE story_id = ? AND title = ?`, ownerID, name).Scan(&rel)
	default:
		if scope == "series" {
			_ = s.DB.QueryRow(`SELECT COALESCE(source_rel, '') FROM series_bible_documents WHERE series_id = ? AND category = ? AND file_name = ?`, ownerID, kind, name).Scan(&rel)
		} else {
			_ = s.DB.QueryRow(`SELECT COALESCE(source_rel, '') FROM story_documents WHERE story_id = ? AND kind = ? AND file_name = ?`, ownerID, kind, name).Scan(&rel)
		}
	}
	return rel
}

func (s *Store) setSourceRelByName(scope string, ownerID int64, kind, name, rel string) {
	switch kind {
	case "character":
		q := `UPDATE character_profiles SET source_rel = ? WHERE series_id = ? AND character_name = ?`
		if scope == "story" {
			q = `UPDATE character_profiles SET source_rel = ? WHERE story_id = ? AND character_name = ?`
		}
		_, _ = s.DB.Exec(q, rel, ownerID, name)
	case "act":
		_, _ = s.DB.Exec(`UPDATE story_acts SET source_rel = ? WHERE story_id = ? AND title = ?`, rel, ownerID, name)
	default:
		if scope == "series" {
			_, _ = s.DB.Exec(`UPDATE series_bible_documents SET source_rel = ? WHERE series_id = ? AND category = ? AND file_name = ?`, rel, ownerID, kind, name)
		} else {
			_, _ = s.DB.Exec(`UPDATE story_documents SET source_rel = ? WHERE story_id = ? AND kind = ? AND file_name = ?`, rel, ownerID, kind, name)
		}
	}
}

func listTextFiles(dir string) ([]diskFile, error) {
	out := make([]diskFile, 0)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != dir && skipSyncDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if skipSyncDir(name) || !textExt(name) {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out = append(out, diskFile{Name: name, Rel: slashRel(rel), Text: string(b)})
		return nil
	})
	return out, err
}

func FolderWatchDirs(root string) []string {
	out := []string{root}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		if path != root && skipSyncDir(d.Name()) {
			return filepath.SkipDir
		}
		if path != root {
			out = append(out, path)
		}
		return nil
	})
	return out
}

func skipSyncDir(name string) bool {
	if name == "." || name == ".." {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch strings.ToLower(name) {
	case "node_modules", "__pycache__", "dist", "build":
		return true
	default:
		return false
	}
}

func textExt(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".markdown", ".txt", ".text":
		return true
	default:
		return false
	}
}

func writeTextFile(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

func mdName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	base = unsafeFileRE.ReplaceAllString(base, " ")
	base = strings.Join(strings.Fields(base), " ")
	if textExt(base) {
		return base
	}
	return base + ".md"
}

func matchKey(name string) string {
	n := strings.TrimSpace(name)
	n = strings.TrimSuffix(n, filepath.Ext(n))
	n = stripChapterPrefix(n)
	return strings.ToLower(strings.TrimSpace(n))
}

func stripChapterPrefix(name string) string {
	n := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	if m := chapterNumRE.FindStringSubmatch(n); m != nil {
		return m[2]
	}
	return n
}

func splitChapterName(name string) (int, string, bool) {
	n := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	m := chapterNumRE.FindStringSubmatch(n)
	if m == nil {
		return 0, n, false
	}
	num, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, m[2], false
	}
	return num, m[2], true
}

func chapterFileName(n int, title string) string {
	t := strings.TrimSpace(title)
	t = unsafeFileRE.ReplaceAllString(t, " ")
	t = strings.Join(strings.Fields(t), " ")
	t = strings.TrimSuffix(t, filepath.Ext(t))
	if t == "" {
		t = "Chapter"
	}
	return fmt.Sprintf("%02d-%s.md", n, t)
}

func chapterSortKey(name string) string {
	if n, title, ok := splitChapterName(name); ok {
		return fmt.Sprintf("%08d-%s", n, strings.ToLower(title))
	}
	return "99999999-" + strings.ToLower(name)
}

func chapterFileSort(f diskFile) string {
	return chapterSortKey(f.Name) + "\t" + f.Rel
}

func slashRel(rel string) string {
	return filepath.ToSlash(rel)
}

func characterImportanceFromRel(rel string) string {
	parts := strings.Split(slashRel(rel), "/")
	if len(parts) < 2 {
		return ""
	}
	switch strings.ToLower(parts[0]) {
	case "main":
		return "Main"
	case "supporting":
		return "Supporting"
	case "minor":
		return "Minor"
	default:
		return ""
	}
}

func relWithBase(rel, base string) string {
	rel = slashRel(rel)
	dir := pathDir(rel)
	if dir == "" || dir == "." {
		return base
	}
	return dir + "/" + base
}

func pathDir(rel string) string {
	i := strings.LastIndex(rel, "/")
	if i <= 0 {
		return ""
	}
	return rel[:i]
}

func uniqueDisplayName(used map[string]bool, rel, base string) string {
	if !used[base] {
		used[base] = true
		return base
	}
	parent := filepath.Base(pathDir(slashRel(rel)))
	if parent == "" || parent == "." {
		parent = "file"
	}
	cand := parent + "-" + base
	if !used[cand] {
		used[cand] = true
		return cand
	}
	for n := 2; ; n++ {
		next := fmt.Sprintf("%s-%d-%s", parent, n, base)
		if !used[next] {
			used[next] = true
			return next
		}
	}
}

type namedRow struct {
	id   int64
	name string
	rel  string
}

func matchFile(rows []namedRow, f diskFile, seen map[int64]bool) int64 {
	for _, r := range rows {
		if seen[r.id] || r.rel == "" {
			continue
		}
		if slashRel(r.rel) == f.Rel {
			seen[r.id] = true
			return r.id
		}
	}
	return matchRow(rows, f.Name, seen)
}

func matchRow(rows []namedRow, name string, seen map[int64]bool) int64 {
	key := matchKey(name)
	for _, r := range rows {
		if seen[r.id] {
			continue
		}
		if matchKey(r.name) == key {
			seen[r.id] = true
			return r.id
		}
	}
	return 0
}

func parseNamedFile(fileName, text string) (name, body string) {
	name = stripChapterPrefix(fileName)
	body = text
	trimmed := strings.TrimLeft(text, "\ufeff \t")
	if strings.HasPrefix(trimmed, "# ") {
		line, rest, _ := strings.Cut(trimmed, "\n")
		h := strings.TrimSpace(strings.TrimPrefix(line, "# "))
		if h != "" {
			name = h
			body = strings.TrimPrefix(rest, "\r")
			body = strings.TrimPrefix(body, "\n")
		}
	}
	return name, body
}

func characterFileText(name, notes string) string {
	if strings.HasPrefix(strings.TrimLeft(notes, " \t"), "# ") {
		return notes
	}
	return "# " + name + "\n\n" + notes
}

func strOr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
