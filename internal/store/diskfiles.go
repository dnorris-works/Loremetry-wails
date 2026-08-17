package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DiskFile struct {
	Dir  string `json:"dir"`
	Name string `json:"name"`
	Path string `json:"path"`
	Text string `json:"text,omitempty"`
}

type DiskFileWrite struct {
	Dir  string `json:"dir"`
	Name string `json:"name"`
	NewName string `json:"new_name"`
	Text string `json:"text"`
}

func ReadDiskFile(dir, name string) (DiskFile, error) {
	path, err := diskFilePath(dir, name)
	if err != nil {
		return DiskFile{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return DiskFile{}, err
	}
	return DiskFile{Dir: dir, Name: filepath.Base(path), Path: path, Text: string(b)}, nil
}

func WriteDiskFile(in DiskFileWrite) (DiskFile, error) {
	name := strings.TrimSpace(in.Name)
	if n := strings.TrimSpace(in.NewName); n != "" && filepath.Base(n) != name {
		next := mdName(n)
		from, err := diskFilePath(in.Dir, name)
		if err != nil {
			return DiskFile{}, err
		}
		to := filepath.Join(in.Dir, next)
		if err := os.Rename(from, to); err != nil && !os.IsNotExist(err) {
			return DiskFile{}, err
		}
		name = next
	}
	path, err := diskFilePath(in.Dir, name)
	if err != nil {
		return DiskFile{}, err
	}
	if err := writeTextFile(path, in.Text); err != nil {
		return DiskFile{}, err
	}
	if shouldRenumberDir(in.Dir) {
		s := &Store{}
		if err := s.renumberChapterFiles(in.Dir); err != nil {
			return DiskFile{}, err
		}
		name = findNameAfterRenumber(in.Dir, name)
		return DiskFile{Dir: in.Dir, Name: name, Path: filepath.Join(in.Dir, name)}, nil
	}
	return DiskFile{Dir: in.Dir, Name: filepath.Base(path), Path: path}, nil
}

func CreateDiskFile(dir, name, text string) (DiskFile, error) {
	if strings.TrimSpace(dir) == "" {
		return DiskFile{}, fmt.Errorf("missing folder")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return DiskFile{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		if shouldRenumberDir(dir) {
			name = "Chapter"
		} else {
			name = nextDocTitle(dir)
		}
	}
	fileName := mdName(name)
	if shouldRenumberDir(dir) {
		files, _ := listTextFileMeta(dir)
		fileName = chapterFileName(len(files)+1, stripChapterPrefix(name))
	}
	path := filepath.Join(dir, fileName)
	if _, err := os.Stat(path); err == nil {
		fileName = filepath.Base(uniqueRel(dir, fileName))
		path = filepath.Join(dir, fileName)
	}
	if err := writeTextFile(path, text); err != nil {
		return DiskFile{}, err
	}
	if shouldRenumberDir(dir) {
		s := &Store{}
		if err := s.renumberChapterFiles(dir); err != nil {
			return DiskFile{}, err
		}
		fileName = findNameAfterRenumber(dir, name)
		path = filepath.Join(dir, fileName)
	}
	return DiskFile{Dir: dir, Name: fileName, Path: path}, nil
}

type IncomingFile struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func CreateDiskFiles(dir string, files []IncomingFile) (DiskFile, error) {
	var last DiskFile
	n := 0
	for _, f := range files {
		if !ImportableFileName(f.Name) {
			continue
		}
		created, err := CreateDiskFile(dir, f.Name, f.Text)
		if err != nil {
			return DiskFile{}, err
		}
		last = created
		n++
	}
	if n == 0 {
		return DiskFile{}, fmt.Errorf("nothing to import")
	}
	return last, nil
}

func ImportableFileName(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range []string{".md", ".txt", ".markdown", ".text", ".fountain", ".json"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return !strings.Contains(filepath.Base(name), ".")
}

func DeleteDiskFile(dir, name string) error {
	path, err := diskFilePath(dir, name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if shouldRenumberDir(dir) {
		s := &Store{}
		return s.renumberChapterFiles(dir)
	}
	return nil
}

func shouldRenumberDir(dir string) bool {
	base := filepath.Base(dir)
	if base == "00_Current-Draft" {
		return true
	}
	files, err := listTextFileMeta(dir)
	if err != nil || len(files) == 0 {
		return false
	}
	n := 0
	for _, f := range files {
		if _, _, ok := splitChapterName(f.Name); ok {
			n++
		}
	}
	return n*2 >= len(files)
}

func findNameAfterRenumber(dir, title string) string {
	rel := findRelAfterRenumber(dir, title, title)
	return filepath.Base(rel)
}

func nextDocTitle(dir string) string {
	used := map[string]bool{}
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !textExt(e.Name()) {
				continue
			}
			stem := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
			used[strings.ToLower(stem)] = true
		}
	}
	if !used["document"] {
		return "Document"
	}
	for n := 2; ; n++ {
		if !used[fmt.Sprintf("document %d", n)] {
			return fmt.Sprintf("Document %d", n)
		}
	}
}

func diskFilePath(dir, name string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", fmt.Errorf("missing folder")
	}
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == ".." || strings.Contains(base, "/") || strings.Contains(base, `\`) {
		return "", fmt.Errorf("invalid name")
	}
	return filepath.Join(dir, base), nil
}
