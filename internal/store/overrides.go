package store

import "strings"

func (s *Store) headerOverride(projectPath, kind string) string {
	if s == nil || s.DB == nil {
		return ""
	}
	var path string
	err := s.DB.QueryRow(`SELECT path FROM header_overrides WHERE project_path = ? AND kind = ?`, projectPath, kind).Scan(&path)
	if err != nil {
		return ""
	}
	return path
}

func (s *Store) ListHeaderOverrides(projectPath string) ([]FolderLink, error) {
	rows, err := s.DB.Query(`SELECT kind, path FROM header_overrides WHERE project_path = ? ORDER BY kind`, projectPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]FolderLink, 0)
	for rows.Next() {
		var l FolderLink
		if err := rows.Scan(&l.Kind, &l.Path); err != nil {
			return nil, err
		}
		l.Scope = projectPath
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) SetHeaderOverride(projectPath, kind, path string) (FolderLink, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return FolderLink{}, nil
	}
	_, err := s.DB.Exec(`
		INSERT INTO header_overrides (project_path, kind, path) VALUES (?, ?, ?)
		ON CONFLICT(project_path, kind) DO UPDATE SET path = excluded.path
	`, projectPath, kind, path)
	if err != nil {
		return FolderLink{}, err
	}
	return FolderLink{Scope: projectPath, Kind: kind, Path: path}, nil
}

func (s *Store) ClearHeaderOverride(projectPath, kind string) error {
	_, err := s.DB.Exec(`DELETE FROM header_overrides WHERE project_path = ? AND kind = ?`, projectPath, kind)
	return err
}

func (s *Store) AllHeaderOverridePaths() ([]string, error) {
	rows, err := s.DB.Query(`SELECT DISTINCT path FROM header_overrides`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
