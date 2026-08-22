package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ChapterSummaryChapter is one manuscript chapter for the Chapter Plot Summary picker.
type ChapterSummaryChapter struct {
	Rel         string `json:"rel"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

const chapterDescMaxRunes = 400

// ListChapterSummaryChapters lists manuscript chapters and seeds missing DB descriptions
// from the first ~2 sentences of each file.
func (s *Store) ListChapterSummaryChapters(projectPath string) ([]ChapterSummaryChapter, error) {
	if s == nil || s.DB == nil {
		return nil, fmt.Errorf("store not ready")
	}
	root := ResolveProjectRoot(projectPath)
	if root == "" {
		return nil, fmt.Errorf("select a book or series first")
	}
	role := MatchAnalysisSources(root).Role("manuscript")
	if !role.Present || len(role.Files) == 0 {
		return []ChapterSummaryChapter{}, nil
	}

	out := make([]ChapterSummaryChapter, 0, len(role.Files))
	for _, f := range role.Files {
		rel := filepath.ToSlash(f.Rel)
		name := strings.TrimSuffix(f.Name, filepath.Ext(f.Name))
		if name == "" {
			name = f.Name
		}
		desc, err := s.chapterDescription(root, rel)
		if err == nil && strings.TrimSpace(desc) != "" {
			out = append(out, ChapterSummaryChapter{Rel: rel, Name: name, Description: desc})
			continue
		}
		b, readErr := os.ReadFile(f.Path)
		if readErr != nil {
			out = append(out, ChapterSummaryChapter{Rel: rel, Name: name, Description: ""})
			continue
		}
		desc = firstTwoSentences(string(b))
		_ = s.upsertChapterDescription(root, rel, desc)
		out = append(out, ChapterSummaryChapter{Rel: rel, Name: name, Description: desc})
	}
	return out, nil
}

func (s *Store) chapterDescription(projectPath, chapterRel string) (string, error) {
	var desc string
	err := s.DB.QueryRow(
		`SELECT description FROM chapter_descriptions WHERE project_path = ? AND chapter_rel = ?`,
		projectPath, chapterRel,
	).Scan(&desc)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return desc, err
}

func (s *Store) upsertChapterDescription(projectPath, chapterRel, description string) error {
	_, err := s.DB.Exec(`
		INSERT INTO chapter_descriptions (project_path, chapter_rel, description, updated_at)
		VALUES (?, ?, ?, datetime('now'))
		ON CONFLICT(project_path, chapter_rel) DO UPDATE SET
			description = excluded.description,
			updated_at = datetime('now')
	`, projectPath, chapterRel, description)
	return err
}

// FilterManuscriptByRels keeps non-manuscript blobs and only the selected manuscript chapters.
// Rels are matched against RoleText.Rel (manuscript-relative slash paths).
func FilterManuscriptByRels(blobs []RoleText, rels []string) []RoleText {
	want := make(map[string]bool, len(rels))
	for _, r := range rels {
		r = filepath.ToSlash(strings.TrimSpace(r))
		if r != "" {
			want[r] = true
		}
	}
	if len(want) == 0 {
		return nil
	}
	out := make([]RoleText, 0, len(blobs))
	for _, b := range blobs {
		if b.Role != "manuscript" {
			out = append(out, b)
			continue
		}
		rel := filepath.ToSlash(b.Rel)
		if want[rel] || want[filepath.Base(rel)] {
			out = append(out, b)
		}
	}
	return out
}

func firstTwoSentences(text string) string {
	sents := splitSentences(text)
	if len(sents) == 0 {
		return ""
	}
	n := 2
	if len(sents) < n {
		n = len(sents)
	}
	joined := strings.TrimSpace(strings.Join(sents[:n], " "))
	runes := []rune(joined)
	if len(runes) > chapterDescMaxRunes {
		return strings.TrimSpace(string(runes[:chapterDescMaxRunes])) + "…"
	}
	return joined
}
