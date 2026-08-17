package store

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

const folderVisKey = "folder_visibility"

type FolderVisibility struct {
	HiddenNames []string                     `json:"hidden_names"`
	ShowHidden  bool                         `json:"show_hidden"`
	Overrides   map[string]map[string]string `json:"overrides"`
}

type FolderOverrideInput struct {
	ProjectPath string `json:"project_path"`
	Key         string `json:"key"`
	Mode        string `json:"mode"`
}

type NameList struct {
	Names []string `json:"names"`
}

type TemplateFolderLists struct {
	Series []string `json:"series"`
	Books  []string `json:"books"`
}

func TemplateFolderNames() TemplateFolderLists {
	return TemplateFolderLists{
		Series: childFolderNames(templates.Series),
		Books:  childFolderNames(templates.Book),
	}
}

func childFolderNames(node templateNode) []string {
	out := make([]string, 0)
	for _, c := range node.Children {
		if c.Type == "folder" || c.Type == "folder_group" {
			if c.Name != "" {
				out = append(out, c.Name)
			}
		}
	}
	return out
}

func (s *Store) GetFolderVisibility() (FolderVisibility, error) {
	out := FolderVisibility{HiddenNames: []string{}, Overrides: map[string]map[string]string{}}
	raw, err := s.settingValue(folderVisKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return out, nil
	}
	if json.Unmarshal([]byte(raw), &out) != nil {
		return FolderVisibility{HiddenNames: []string{}, Overrides: map[string]map[string]string{}}, nil
	}
	if out.HiddenNames == nil {
		out.HiddenNames = []string{}
	}
	if out.Overrides == nil {
		out.Overrides = map[string]map[string]string{}
	}
	return out, nil
}

func (s *Store) SetHiddenFolderNames(names []string) (FolderVisibility, error) {
	vis, err := s.GetFolderVisibility()
	if err != nil {
		return FolderVisibility{}, err
	}
	clean := make([]string, 0, len(names))
	seen := map[string]bool{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		clean = append(clean, n)
	}
	vis.HiddenNames = clean
	return s.saveFolderVisibility(vis)
}

func (s *Store) SetShowHiddenFolders(on bool) (FolderVisibility, error) {
	vis, err := s.GetFolderVisibility()
	if err != nil {
		return FolderVisibility{}, err
	}
	vis.ShowHidden = on
	return s.saveFolderVisibility(vis)
}

func (s *Store) SetFolderOverride(in FolderOverrideInput) (FolderVisibility, error) {
	vis, err := s.GetFolderVisibility()
	if err != nil {
		return FolderVisibility{}, err
	}
	project := strings.TrimSpace(in.ProjectPath)
	key := strings.TrimSpace(in.Key)
	if project == "" || key == "" {
		return vis, nil
	}
	mode := strings.ToLower(strings.TrimSpace(in.Mode))
	if vis.Overrides[project] == nil {
		vis.Overrides[project] = map[string]string{}
	}
	if mode == "hide" || mode == "show" {
		vis.Overrides[project][key] = mode
	} else {
		delete(vis.Overrides[project], key)
		if len(vis.Overrides[project]) == 0 {
			delete(vis.Overrides, project)
		}
	}
	return s.saveFolderVisibility(vis)
}

func (s *Store) saveFolderVisibility(vis FolderVisibility) (FolderVisibility, error) {
	b, err := json.Marshal(vis)
	if err != nil {
		return FolderVisibility{}, err
	}
	if _, err := s.PutSetting(folderVisKey, string(b)); err != nil {
		return FolderVisibility{}, err
	}
	return vis, nil
}

func (v FolderVisibility) IsHidden(projectPath, name, rel string) bool {
	mode := v.override(projectPath, rel)
	if mode == "" {
		mode = v.override(projectPath, name)
	}
	if mode == "show" {
		return false
	}
	if mode == "hide" {
		return true
	}
	return v.hasName(name) || v.hasName(rel)
}

func (v FolderVisibility) override(projectPath, key string) string {
	if projectPath == "" || key == "" || v.Overrides == nil {
		return ""
	}
	return v.Overrides[projectPath][key]
}

func (v FolderVisibility) hasName(s string) bool {
	for _, n := range v.HiddenNames {
		if n == s {
			return true
		}
	}
	return false
}

func FilterWritingTree(tree WritingTree, vis FolderVisibility) WritingTree {
	for i := range tree.Pens {
		tree.Pens[i] = filterPen(tree.Pens[i], vis)
	}
	return tree
}

func filterPen(pen TreePen, vis FolderVisibility) TreePen {
	for i := range pen.Series {
		pen.Series[i].Tree = filterProjectTree(pen.Series[i].Path, pen.Series[i].Tree, vis)
		for j := range pen.Series[i].Books {
			pen.Series[i].Books[j].Tree = filterProjectTree(pen.Series[i].Books[j].Path, pen.Series[i].Books[j].Tree, vis)
		}
	}
	for i := range pen.Books {
		pen.Books[i].Tree = filterProjectTree(pen.Books[i].Path, pen.Books[i].Tree, vis)
	}
	return pen
}

func filterProjectTree(projectPath string, node DirNode, vis FolderVisibility) DirNode {
	kept := make([]DirNode, 0, len(node.Folders))
	for _, child := range node.Folders {
		if next, ok := filterFolder(projectPath, "", child, vis); ok {
			kept = append(kept, next)
		}
	}
	node.Folders = kept
	return recount(node)
}

func filterFolder(projectPath, parentRel string, n DirNode, vis FolderVisibility) (DirNode, bool) {
	rel := n.Name
	if parentRel != "" {
		rel = parentRel + "/" + n.Name
	}
	rel = filepath.ToSlash(rel)
	hidden := vis.IsHidden(projectPath, n.Name, rel)
	if hidden && !vis.ShowHidden {
		return n, false
	}
	n.Hidden = hidden
	kids := make([]DirNode, 0, len(n.Folders))
	for _, c := range n.Folders {
		if next, ok := filterFolder(projectPath, rel, c, vis); ok {
			kids = append(kids, next)
		}
	}
	n.Folders = kids
	return recount(n), true
}

func recount(n DirNode) DirNode {
	n.Count = len(n.Files)
	for i := range n.Folders {
		n.Count += n.Folders[i].Count
	}
	n.Label = fmt.Sprintf("%s (%d)", n.Name, n.Count)
	return n
}
