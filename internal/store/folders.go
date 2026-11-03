package store

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *Store) ListFolderLinks(scope string, ownerID int64) ([]FolderLink, error) {
	if err := s.ownsFolderOwner(scope, ownerID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`
		SELECT scope, owner_id, kind, path FROM folder_links
		WHERE scope = ? AND owner_id = ? ORDER BY kind`, scope, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]FolderLink, 0)
	for rows.Next() {
		var l FolderLink
		if err := rows.Scan(&l.Scope, &l.OwnerID, &l.Kind, &l.Path); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) AllFolderLinks() ([]FolderLink, error) {
	rows, err := s.DB.Query(`SELECT scope, owner_id, kind, path FROM folder_links ORDER BY scope, owner_id, kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]FolderLink, 0)
	for rows.Next() {
		var l FolderLink
		if err := rows.Scan(&l.Scope, &l.OwnerID, &l.Kind, &l.Path); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) SetFolderLink(in FolderLinkInput) (FolderLink, error) {
	if err := s.ownsFolderOwner(in.Scope, in.OwnerID); err != nil {
		return FolderLink{}, err
	}
	if !validFolderKind(in.Kind) {
		return FolderLink{}, fmt.Errorf("unknown header")
	}
	info, err := os.Stat(in.Path)
	if err != nil || !info.IsDir() {
		return FolderLink{}, fmt.Errorf("folder not found")
	}
	path := filepath.Clean(in.Path)
	_, err = s.DB.Exec(`
		INSERT INTO folder_links (scope, owner_id, kind, path)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(scope, owner_id, kind) DO UPDATE SET path = excluded.path`,
		in.Scope, in.OwnerID, in.Kind, path)
	if err != nil {
		return FolderLink{}, err
	}
	if in.Kind != "root" {
		if err := s.exportIfFolderEmpty(in.Scope, in.OwnerID, in.Kind, path); err != nil {
			return FolderLink{}, err
		}
		if _, err := s.SyncFolder(in.Scope, in.OwnerID, in.Kind); err != nil {
			return FolderLink{}, err
		}
	}
	return FolderLink{Scope: in.Scope, OwnerID: in.OwnerID, Kind: in.Kind, Path: path}, nil
}

func (s *Store) ClearFolderLink(scope string, ownerID int64, kind string) (DeletedResult, error) {
	if err := s.ownsFolderOwner(scope, ownerID); err != nil {
		return DeletedResult{}, err
	}
	_, err := s.DB.Exec(`DELETE FROM folder_links WHERE scope = ? AND owner_id = ? AND kind = ?`, scope, ownerID, kind)
	if err != nil {
		return DeletedResult{}, err
	}
	return DeletedResult{Deleted: true}, nil
}

func (s *Store) folderPath(scope string, ownerID int64, kind string) string {
	var path string
	err := s.DB.QueryRow(`SELECT path FROM folder_links WHERE scope = ? AND owner_id = ? AND kind = ?`,
		scope, ownerID, kind).Scan(&path)
	if err != nil {
		return ""
	}
	return path
}

func (s *Store) ownsFolderOwner(scope string, ownerID int64) error {
	switch scope {
	case "series":
		return s.ownsSeries(ownerID)
	case "story":
		return s.ownsStory(ownerID)
	default:
		return fmt.Errorf("scope must be series or story")
	}
}

func validFolderKind(kind string) bool {
	switch kind {
	case "bible", "location", "research", "character", "act", "chapter", "root":
		return true
	default:
		return false
	}
}
