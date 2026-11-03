package store

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ListPens(writingRoot string, series []Series, stories []Story) []Pen {
	seen := map[string]bool{}
	out := make([]Pen, 0)
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, Pen{Name: name})
	}
	if writingRoot != "" {
		entries, err := os.ReadDir(writingRoot)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() || skipSyncDir(e.Name()) {
					continue
				}
				add(e.Name())
			}
		}
	}
	for _, s := range series {
		add(s.PenName)
	}
	for _, st := range stories {
		if st.SeriesID == nil {
			add(st.PenName)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func penKey(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.ReplaceAll(n, "-", " ")
	n = strings.ReplaceAll(n, "_", " ")
	return strings.Join(strings.Fields(n), " ")
}

func FindPenDir(writingRoot, penName string) string {
	if writingRoot == "" || strings.TrimSpace(penName) == "" {
		return ""
	}
	exact := filepath.Join(writingRoot, strings.TrimSpace(penName))
	if info, err := os.Stat(exact); err == nil && info.IsDir() {
		return exact
	}
	want := penKey(penName)
	entries, err := os.ReadDir(writingRoot)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() || skipSyncDir(e.Name()) {
			continue
		}
		if penKey(e.Name()) == want {
			return filepath.Join(writingRoot, e.Name())
		}
	}
	return ""
}

func EnsurePenDir(writingRoot, penName string) (string, error) {
	if found := FindPenDir(writingRoot, penName); found != "" {
		return found, nil
	}
	name := safeFolderName(penName)
	path := filepath.Join(writingRoot, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", err
	}
	return path, nil
}
