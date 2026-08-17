package store

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

//go:embed project-template.json
var projectTemplateJSON []byte

type templateNode struct {
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Children   []templateNode `json:"children"`
	SeriesOnly bool           `json:"series_only"`
}

type projectTemplates struct {
	Book   templateNode `json:"book_template"`
	Series templateNode `json:"series_template"`
}

var templates projectTemplates

func init() {
	if err := json.Unmarshal(projectTemplateJSON, &templates); err != nil {
		panic(err)
	}
}

var templateToken = regexp.MustCompile(`\{\{(\w+)\}\}`)
var folderUnsafe = regexp.MustCompile(`[\\/:*?"<>|]`)

func ApplySeriesTemplate(parent, seriesName string) (string, error) {
	vars := map[string]string{"SeriesName": safeFolderName(seriesName)}
	return applyRoot(parent, templates.Series, vars, true)
}

func ApplyBookTemplate(parent, bookTitle, seriesName, section string) (string, error) {
	inSeries := seriesName != ""
	title := safeFolderName(bookTitle)
	name := title
	if inSeries {
		name = fmt.Sprintf("%02d_%s", nextBookIndex(parent), title)
	}
	node := templates.Book
	node.Name = name
	vars := map[string]string{"BookTitle": title, "SeriesName": safeFolderName(seriesName), "DraftSection": NormalizeDraftSection(section)}
	return applyRoot(parent, node, vars, inSeries)
}

func applyRoot(parent string, node templateNode, vars map[string]string, inSeries bool) (string, error) {
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	root, err := applyNode(parent, node, vars, inSeries)
	if err != nil {
		return "", err
	}
	return root, nil
}

func applyNode(parent string, node templateNode, vars map[string]string, inSeries bool) (string, error) {
	if node.SeriesOnly && !inSeries {
		return "", nil
	}
	name := expandTemplate(node.Name, vars)
	if name == "" {
		return "", nil
	}
	path := filepath.Join(parent, name)
	switch node.Type {
	case "folder", "folder_group":
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", err
		}
		for _, child := range node.Children {
			if _, err := applyNode(path, child, vars, inSeries); err != nil {
				return "", err
			}
		}
		return path, nil
	case "file":
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
			return "", err
		}
		return path, nil
	default:
		return "", fmt.Errorf("unknown template type %q", node.Type)
	}
}

func expandTemplate(s string, vars map[string]string) string {
	return templateToken.ReplaceAllStringFunc(s, func(tok string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(tok, "{{"), "}}")
		if v, ok := vars[key]; ok {
			return v
		}
		return tok
	})
}

func safeFolderName(name string) string {
	n := strings.TrimSpace(name)
	n = folderUnsafe.ReplaceAllString(n, " ")
	n = strings.Join(strings.Fields(n), " ")
	if n == "" {
		return "Untitled"
	}
	return n
}

func nextBookIndex(booksDir string) int {
	entries, err := os.ReadDir(booksDir)
	if err != nil {
		return 1
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n := e.Name()
		if len(n) < 3 || n[2] != '_' {
			continue
		}
		i, err := strconv.Atoi(n[:2])
		if err != nil {
			continue
		}
		if i > max {
			max = i
		}
	}
	return max + 1
}

func (s *Store) SetRootLink(scope string, ownerID int64, path string) error {
	_, err := s.SetFolderLink(FolderLinkInput{Scope: scope, OwnerID: ownerID, Kind: "root", Path: path})
	return err
}

func (s *Store) RootPath(scope string, ownerID int64) string {
	return s.folderPath(scope, ownerID, "root")
}

func (s *Store) LinkProjectHeaders(scope string, ownerID int64, root string) error {
	dirs := storyHeaderDirs
	if scope == "series" {
		dirs = seriesHeaderDirs
	}
	for kind, rel := range dirs {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(p, 0o755); err != nil {
			return err
		}
		if _, err := s.SetFolderLink(FolderLinkInput{Scope: scope, OwnerID: ownerID, Kind: kind, Path: p}); err != nil {
			return err
		}
	}
	return nil
}

var seriesHeaderDirs = map[string]string{
	"bible":     "00_Series-Bible",
	"character": "00_Series-Bible/Characters",
	"location":  "00_Series-Bible/Locations",
	"research":  "00_Series-Bible/World",
}

var storyHeaderDirs = map[string]string{
	"chapter":   "01_Manuscript/00_Current-Draft",
	"act":       "01_Manuscript/00_Current-Draft",
	"character": "02_Story-Elements/Characters",
	"location":  "02_Story-Elements/Locations",
	"research":  "04_Research",
	"bible":     "02_Story-Elements/World-Notes",
}
