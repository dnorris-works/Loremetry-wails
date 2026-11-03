package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxHashBytes = 256 << 10 // 256 KB — larger files use size+modtime

func (s *Store) SyncDiskHashes(root string) ([]string, error) {
	root = filepath.Clean(strings.TrimSpace(root))
	if root == "" {
		return nil, nil
	}
	next, err := scanDiskHashes(root)
	if err != nil {
		return nil, err
	}
	prev, err := s.loadDiskHashes(root)
	if err != nil {
		return nil, err
	}
	if hashesEqual(prev, next) {
		return nil, nil
	}
	changed := diffHashes(prev, next)
	if err := s.replaceDiskHashes(root, next); err != nil {
		return nil, err
	}
	return changed, nil
}

func scanDiskHashes(root string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if path != root && skipSyncDir(name) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			out[rel] = "dir"
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		h, err := hashFile(path)
		if err != nil {
			return nil
		}
		out[rel] = h
		return nil
	})
	return out, err
}

func hashFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > maxHashBytes {
		sum := sha256.Sum256(fmt.Appendf(nil, "%d:%d:%s", info.Size(), info.ModTime().UnixNano(), filepath.Base(path)))
		return hex.EncodeToString(sum[:]), nil
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (s *Store) loadDiskHashes(root string) (map[string]string, error) {
	rows, err := s.DB.Query(`SELECT rel, hash FROM disk_hashes WHERE root = ?`, root)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var rel, hash string
		if err := rows.Scan(&rel, &hash); err != nil {
			return nil, err
		}
		out[rel] = hash
	}
	return out, rows.Err()
}

func (s *Store) replaceDiskHashes(root string, next map[string]string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM disk_hashes WHERE root = ?`, root); err != nil {
		return err
	}
	st, err := tx.Prepare(`INSERT INTO disk_hashes (root, rel, hash) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer st.Close()
	for rel, hash := range next {
		if _, err := st.Exec(root, rel, hash); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func hashesEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func diffHashes(prev, next map[string]string) []string {
	out := make([]string, 0)
	for k, v := range next {
		if prev[k] != v {
			out = append(out, k)
		}
	}
	for k := range prev {
		if _, ok := next[k]; !ok {
			out = append(out, k)
		}
	}
	return out
}
