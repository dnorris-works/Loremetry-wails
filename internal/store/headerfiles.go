package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type HeaderFile struct {
	Rel    string `json:"rel"`
	Name   string `json:"name"`
	Folder string `json:"folder,omitempty"`
}

type HeaderList struct {
	Folder  string       `json:"folder"`
	Missing bool         `json:"missing"`
	Files   []HeaderFile `json:"files"`
}

type HeaderFileContent struct {
	Rel  string `json:"rel"`
	Name string `json:"name"`
	Text string `json:"text"`
}

type HeaderFileRef struct {
	ProjectPath string `json:"project_path"`
	ProjectKind string `json:"project_kind"`
	Kind        string `json:"kind"`
	Rel         string `json:"rel"`
}

type HeaderFileWrite struct {
	ProjectPath string `json:"project_path"`
	ProjectKind string `json:"project_kind"`
	Kind        string `json:"kind"`
	Rel         string `json:"rel"`
	Name        string `json:"name"`
	Text        string `json:"text"`
}

type HeaderPlaceInput struct {
	ProjectPath string   `json:"project_path"`
	ProjectKind string   `json:"project_kind"`
	Kind        string   `json:"kind"`
	Rels        []string `json:"rels"`
}

func HeaderDir(s *Store, projectPath, projectKind, kind string) (string, error) {
	if strings.TrimSpace(projectPath) == "" {
		return "", fmt.Errorf("missing project folder")
	}
	if s != nil {
		if over := s.headerOverride(projectPath, kind); over != "" {
			return over, nil
		}
	}
	rel := defaultHeaderRel(projectKind, kind)
	if rel == "" {
		return "", fmt.Errorf("unknown header")
	}
	return filepath.Join(projectPath, filepath.FromSlash(rel)), nil
}

func defaultHeaderRel(projectKind, kind string) string {
	if projectKind == "series" {
		if kind == "bible" {
			return "00_Series-Bible"
		}
		return seriesHeaderDirs[kind]
	}
	return storyHeaderDirs[kind]
}

func (s *Store) ListHeaderFiles(projectPath, projectKind, kind string) (HeaderList, error) {
	dir, err := HeaderDir(s, projectPath, projectKind, kind)
	if err != nil {
		return HeaderList{}, err
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return HeaderList{Folder: dir, Missing: true, Files: []HeaderFile{}}, nil
	}
	files, err := listTextFileMeta(dir)
	if err != nil {
		return HeaderList{}, err
	}
	if projectKind == "series" && kind == "bible" {
		top := files[:0]
		for _, f := range files {
			if !strings.Contains(f.Rel, "/") {
				top = append(top, f)
			}
		}
		files = top
	}
	if kind == "chapter" {
		sort.Slice(files, func(i, j int) bool {
			return chapterFileSort(files[i]) < chapterFileSort(files[j])
		})
	} else {
		sort.Slice(files, func(i, j int) bool {
			return strings.ToLower(files[i].Rel) < strings.ToLower(files[j].Rel)
		})
	}
	out := HeaderList{Folder: dir, Files: make([]HeaderFile, 0, len(files))}
	for _, f := range files {
		out.Files = append(out.Files, HeaderFile{Rel: f.Rel, Name: f.Name, Folder: dir})
	}
	return out, nil
}

func (s *Store) ReadHeaderFile(in HeaderFileRef) (HeaderFileContent, error) {
	dir, err := HeaderDir(s, in.ProjectPath, in.ProjectKind, in.Kind)
	if err != nil {
		return HeaderFileContent{}, err
	}
	rel, err := safeRel(in.Rel)
	if err != nil {
		return HeaderFileContent{}, err
	}
	path := filepath.Join(dir, filepath.FromSlash(rel))
	b, err := os.ReadFile(path)
	if err != nil {
		return HeaderFileContent{}, err
	}
	return HeaderFileContent{Rel: slashRel(rel), Name: filepath.Base(rel), Text: string(b)}, nil
}

func (s *Store) WriteHeaderFile(in HeaderFileWrite) (HeaderFile, error) {
	dir, err := HeaderDir(s, in.ProjectPath, in.ProjectKind, in.Kind)
	if err != nil {
		return HeaderFile{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return HeaderFile{}, err
	}
	rel := strings.TrimSpace(in.Rel)
	if rel == "" {
		return HeaderFile{}, fmt.Errorf("missing file")
	}
	rel, err = safeRel(rel)
	if err != nil {
		return HeaderFile{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name != "" && filepath.Base(rel) != name && filepath.Base(rel) != mdName(name) {
		nextRel := relWithBase(rel, mdName(name))
		if err := os.Rename(filepath.Join(dir, filepath.FromSlash(rel)), filepath.Join(dir, filepath.FromSlash(nextRel))); err != nil && !os.IsNotExist(err) {
			return HeaderFile{}, err
		}
		rel = nextRel
	}
	if err := writeTextFile(filepath.Join(dir, filepath.FromSlash(rel)), in.Text); err != nil {
		return HeaderFile{}, err
	}
	if in.Kind == "chapter" {
		if err := s.renumberChapterFiles(dir); err != nil {
			return HeaderFile{}, err
		}
		rel = findRelAfterRenumber(dir, rel, name)
	}
	return HeaderFile{Rel: slashRel(rel), Name: filepath.Base(rel), Folder: dir}, nil
}

func (s *Store) CreateHeaderFile(in HeaderFileWrite) (HeaderFile, error) {
	dir, err := HeaderDir(s, in.ProjectPath, in.ProjectKind, in.Kind)
	if err != nil {
		return HeaderFile{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return HeaderFile{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = in.Kind
	}
	fileName := mdName(name)
	if in.Kind == "chapter" {
		files, _ := listTextFileMeta(dir)
		fileName = chapterFileName(len(files)+1, stripChapterPrefix(name))
	}
	rel := fileName
	if strings.TrimSpace(in.Rel) != "" {
		baseRel, err := safeRel(in.Rel)
		if err != nil {
			return HeaderFile{}, err
		}
		rel = relWithBase(baseRel, fileName)
	}
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if _, err := os.Stat(path); err == nil {
		rel = uniqueRel(dir, rel)
		path = filepath.Join(dir, filepath.FromSlash(rel))
	}
	if err := writeTextFile(path, in.Text); err != nil {
		return HeaderFile{}, err
	}
	if in.Kind == "chapter" {
		if err := s.renumberChapterFiles(dir); err != nil {
			return HeaderFile{}, err
		}
		rel = findRelAfterRenumber(dir, rel, name)
	}
	return HeaderFile{Rel: slashRel(rel), Name: filepath.Base(rel), Folder: dir}, nil
}

func (s *Store) DeleteHeaderFile(in HeaderFileRef) error {
	dir, err := HeaderDir(s, in.ProjectPath, in.ProjectKind, in.Kind)
	if err != nil {
		return err
	}
	rel, err := safeRel(in.Rel)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dir, filepath.FromSlash(rel))); err != nil && !os.IsNotExist(err) {
		return err
	}
	if in.Kind == "chapter" {
		return s.renumberChapterFiles(dir)
	}
	return nil
}

func (s *Store) PlaceHeaderFiles(in HeaderPlaceInput) error {
	if in.Kind != "chapter" {
		return nil
	}
	dir, err := HeaderDir(s, in.ProjectPath, in.ProjectKind, in.Kind)
	if err != nil {
		return err
	}
	files, err := listTextFileMeta(dir)
	if err != nil {
		return err
	}
	byRel := map[string]diskFile{}
	for _, f := range files {
		byRel[slashRel(f.Rel)] = f
	}
	ordered := make([]diskFile, 0, len(in.Rels))
	seen := map[string]bool{}
	for _, rel := range in.Rels {
		rel = slashRel(rel)
		if f, ok := byRel[rel]; ok {
			ordered = append(ordered, f)
			seen[rel] = true
		}
	}
	for _, f := range files {
		if !seen[slashRel(f.Rel)] {
			ordered = append(ordered, f)
		}
	}
	return s.rewriteNumberedFiles(dir, ordered, func(i int, f diskFile) string {
		return chapterFileName(i+1, stripChapterPrefix(f.Name))
	})
}

func (s *Store) renumberChapterFiles(dir string) error {
	files, err := listTextFileMeta(dir)
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool {
		return chapterFileSort(files[i]) < chapterFileSort(files[j])
	})
	return s.rewriteNumberedFiles(dir, files, func(i int, f diskFile) string {
		return chapterFileName(i+1, stripChapterPrefix(f.Name))
	})
}

func findRelAfterRenumber(dir, oldRel, title string) string {
	files, err := listTextFileMeta(dir)
	if err != nil {
		return slashRel(oldRel)
	}
	wantTitle := strings.ToLower(stripChapterPrefix(title))
	if wantTitle == "" {
		wantTitle = strings.ToLower(stripChapterPrefix(filepath.Base(oldRel)))
	}
	dirPart := slashRel(filepath.Dir(oldRel))
	for _, f := range files {
		if dirPart != "." && slashRel(filepath.Dir(f.Rel)) != dirPart {
			continue
		}
		if strings.ToLower(stripChapterPrefix(f.Name)) == wantTitle {
			return slashRel(f.Rel)
		}
	}
	if len(files) > 0 {
		return slashRel(files[len(files)-1].Rel)
	}
	return slashRel(oldRel)
}

func uniqueRel(dir, rel string) string {
	rel = slashRel(rel)
	ext := filepath.Ext(rel)
	stem := strings.TrimSuffix(rel, ext)
	for n := 2; ; n++ {
		next := fmt.Sprintf("%s-%d%s", stem, n, ext)
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(next))); err != nil {
			return next
		}
	}
}

func safeRel(rel string) (string, error) {
	rel = slashRel(strings.TrimSpace(rel))
	if rel == "" || rel == "." || strings.HasPrefix(rel, "/") || strings.Contains(rel, "..") {
		return "", fmt.Errorf("invalid path")
	}
	return rel, nil
}

func listTextFileMeta(dir string) ([]diskFile, error) {
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
		out = append(out, diskFile{Name: name, Rel: slashRel(rel)})
		return nil
	})
	return out, err
}
