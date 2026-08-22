package store

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type AnalysisFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Rel  string `json:"rel"`
}

type AnalysisRoleMatch struct {
	Role    string         `json:"role"`
	Rel     string         `json:"rel"`
	Path    string         `json:"path"`
	Present bool           `json:"present"`
	Files   []AnalysisFile `json:"files"`
}

type AnalysisSources struct {
	ProjectPath string              `json:"project_path"`
	Kind        string              `json:"kind"`
	Roles       []AnalysisRoleMatch `json:"roles"`
}

type RoleText struct {
	Role string `json:"role"`
	Rel  string `json:"rel"`
	Name string `json:"name"`
	Text string `json:"text"`
}

func ResolveProjectRoot(start string) string {
	p := filepath.Clean(strings.TrimSpace(start))
	if p == "" || p == "." {
		return ""
	}
	info, err := os.Stat(p)
	if err == nil && !info.IsDir() {
		p = filepath.Dir(p)
	}
	cur := p
	for {
		if looksBook(cur) || looksSeries(cur) {
			return cur
		}
		next := filepath.Dir(cur)
		if next == cur {
			break
		}
		cur = next
	}
	return p
}

const maxSourceBytes = 400_000

func CollectNeededText(projectPath string, needs []string) []RoleText {
	src := MatchAnalysisSources(projectPath)
	var out []RoleText
	used := 0
	for _, need := range needs {
		role := src.Role(need)
		for _, f := range role.Files {
			if used >= maxSourceBytes {
				return out
			}
			b, err := os.ReadFile(f.Path)
			if err != nil {
				continue
			}
			if used+len(b) > maxSourceBytes {
				b = b[:maxSourceBytes-used]
			}
			out = append(out, RoleText{Role: need, Rel: f.Rel, Name: f.Name, Text: string(b)})
			used += len(b)
		}
	}
	return out
}

type analysisRoleSpec struct {
	role string
	rel  string
	file bool
}

func bookAnalysisRoles() []analysisRoleSpec {
	return []analysisRoleSpec{
		{role: "manuscript", rel: "01_Manuscript/01_Chapters"},
		{role: "characters", rel: "02_Story-Elements/Characters"},
		{role: "locations", rel: "02_Story-Elements/Locations"},
		{role: "themes", rel: "02_Story-Elements/Themes"},
		{role: "bible", rel: "02_Story-Elements/World-Notes"},
		{role: "plot", rel: "03_Plot"},
		{role: "continuity", rel: "02_Story-Elements/Continuity-Notes.md", file: true},
		{role: "blurb", rel: "07_Marketing/Blurb.md", file: true},
	}
}

func seriesAnalysisRoles() []analysisRoleSpec {
	return []analysisRoleSpec{
		{role: "bible", rel: "00_Series-Bible"},
		{role: "characters", rel: "00_Series-Bible/Characters"},
		{role: "locations", rel: "00_Series-Bible/Locations"},
		{role: "themes", rel: "00_Series-Bible/Premise-and-Themes.md", file: true},
		{role: "world", rel: "00_Series-Bible/World"},
	}
}

func MatchAnalysisSources(projectPath string) AnalysisSources {
	root := filepath.Clean(strings.TrimSpace(projectPath))
	out := AnalysisSources{ProjectPath: root, Roles: []AnalysisRoleMatch{}}
	if root == "" {
		return out
	}
	specs := bookAnalysisRoles()
	out.Kind = "book"
	if looksSeries(root) {
		out.Kind = "series"
		specs = seriesAnalysisRoles()
	}
	for _, spec := range specs {
		out.Roles = append(out.Roles, matchRole(root, spec))
	}
	return out
}

func (s AnalysisSources) Role(name string) AnalysisRoleMatch {
	for _, r := range s.Roles {
		if r.Role == name {
			return r
		}
	}
	return AnalysisRoleMatch{Role: name, Files: []AnalysisFile{}}
}

func matchRole(root string, spec analysisRoleSpec) AnalysisRoleMatch {
	m := AnalysisRoleMatch{Role: spec.role, Rel: spec.rel, Files: []AnalysisFile{}}
	if spec.file {
		path := resolveFile(root, spec.rel)
		if path == "" {
			return m
		}
		m.Path = path
		m.Present = true
		m.Files = []AnalysisFile{{Name: filepath.Base(path), Path: path, Rel: filepath.Base(path)}}
		return m
	}
	dir := resolveSubdir(root, spec.rel)
	if dir == "" {
		return m
	}
	m.Path = dir
	m.Present = true
	m.Files = collectRoleFiles(dir, spec.role)
	return m
}

func collectRoleFiles(dir, role string) []AnalysisFile {
	var out []AnalysisFile
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if skipSyncDir(name) && path != dir {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if role == "manuscript" && skipDraftNoise(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if !textExt(name) {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		out = append(out, AnalysisFile{
			Name: name,
			Path: path,
			Rel:  filepath.ToSlash(rel),
		})
		return nil
	})
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Rel) < strings.ToLower(out[j].Rel)
	})
	return out
}

func skipDraftNoise(name string) bool {
	switch strings.ToLower(name) {
	case "01_compiled", "02_archive", "03_deleted-scenes":
		return true
	default:
		return false
	}
}

func resolveSubdir(root, relative string) string {
	relative = strings.Trim(strings.ReplaceAll(relative, "\\", "/"), "/")
	if relative == "" {
		return ""
	}
	exact := filepath.Join(root, filepath.FromSlash(relative))
	if info, err := os.Stat(exact); err == nil && info.IsDir() {
		return exact
	}
	current := root
	for _, seg := range strings.Split(relative, "/") {
		if seg == "" {
			continue
		}
		next := filepath.Join(current, seg)
		if info, err := os.Stat(next); err == nil && info.IsDir() {
			current = next
			continue
		}
		found := findDirIgnoreCase(current, seg)
		if found == "" {
			return ""
		}
		current = found
	}
	if info, err := os.Stat(current); err == nil && info.IsDir() {
		return current
	}
	return ""
}

func resolveFile(root, relative string) string {
	relative = strings.Trim(strings.ReplaceAll(relative, "\\", "/"), "/")
	dirRel, name := filepath.Split(relative)
	dirRel = strings.Trim(dirRel, "/")
	parent := root
	if dirRel != "" {
		parent = resolveSubdir(root, dirRel)
		if parent == "" {
			return ""
		}
	}
	exact := filepath.Join(parent, name)
	if info, err := os.Stat(exact); err == nil && info.Mode().IsRegular() {
		return exact
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(parent, e.Name())
		}
	}
	return ""
}

func findDirIgnoreCase(parent, name string) string {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(parent, e.Name())
		}
	}
	return ""
}
