package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const draftSectionKey = "draft_section"

var draftSectionFolder = regexp.MustCompile(`(?i)^(part|act)[-_](.+)$`)

func NormalizeDraftSection(s string) string {
	if strings.EqualFold(strings.TrimSpace(s), "part") {
		return "Part"
	}
	return "Act"
}

func (s *Store) GetDraftSection() string {
	got, err := s.GetSetting(draftSectionKey)
	if err != nil {
		return "Act"
	}
	return NormalizeDraftSection(got.Value)
}

func (s *Store) SetDraftSection(value string) (string, error) {
	v := NormalizeDraftSection(value)
	if _, err := s.PutSetting(draftSectionKey, v); err != nil {
		return "", err
	}
	return v, nil
}

func RenameDraftSections(root, to string) (int, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return 0, fmt.Errorf("writing folder is not set")
	}
	want := NormalizeDraftSection(to)
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && skipSyncDir(d.Name()) {
			return filepath.SkipDir
		}
		if draftSectionFolder.MatchString(d.Name()) {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})
	n := 0
	for _, path := range dirs {
		base := filepath.Base(path)
		m := draftSectionFolder.FindStringSubmatch(base)
		if m == nil {
			continue
		}
		if strings.EqualFold(m[1], want) {
			continue
		}
		next := filepath.Join(filepath.Dir(path), want+"-"+m[2])
		if _, err := os.Stat(next); err == nil {
			continue
		}
		if err := os.Rename(path, next); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
