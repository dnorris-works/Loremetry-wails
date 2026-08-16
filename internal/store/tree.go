package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var bookFolderPrefix = regexp.MustCompile(`^\d+_`)

type TreeProblem struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type TreeBook struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

type TreeSeries struct {
	Path     string        `json:"path"`
	Name     string        `json:"name"`
	Books    []TreeBook    `json:"books"`
	Problems []TreeProblem `json:"problems"`
}

type TreePen struct {
	Name     string        `json:"name"`
	Path     string        `json:"path"`
	Series   []TreeSeries  `json:"series"`
	Books    []TreeBook    `json:"books"`
	Problems []TreeProblem `json:"problems"`
}

type WritingTree struct {
	Root     string        `json:"root"`
	Pens     []TreePen     `json:"pens"`
	Problems []TreeProblem `json:"problems"`
}

func ListWritingTree(root string) WritingTree {
	out := WritingTree{Root: root, Pens: []TreePen{}, Problems: []TreeProblem{}}
	if strings.TrimSpace(root) == "" {
		out.Problems = append(out.Problems, TreeProblem{Message: "Writing folder is not set. Choose one in Settings."})
		return out
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		out.Problems = append(out.Problems, TreeProblem{Path: root, Message: "Writing folder is missing or is not a directory."})
		return out
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		out.Problems = append(out.Problems, TreeProblem{Path: root, Message: err.Error()})
		return out
	}
	for _, e := range entries {
		if !e.IsDir() || skipSyncDir(e.Name()) {
			continue
		}
		out.Pens = append(out.Pens, scanPen(filepath.Join(root, e.Name()), e.Name()))
	}
	sort.Slice(out.Pens, func(i, j int) bool {
		return strings.ToLower(out.Pens[i].Name) < strings.ToLower(out.Pens[j].Name)
	})
	return out
}

func scanPen(path, name string) TreePen {
	pen := TreePen{Name: name, Path: path, Series: []TreeSeries{}, Books: []TreeBook{}, Problems: []TreeProblem{}}
	entries, err := os.ReadDir(path)
	if err != nil {
		pen.Problems = append(pen.Problems, TreeProblem{Path: path, Message: err.Error()})
		return pen
	}
	var seriesPaths []string
	for _, e := range entries {
		if !e.IsDir() || skipSyncDir(e.Name()) {
			continue
		}
		child := filepath.Join(path, e.Name())
		if looksSeries(child) {
			seriesPaths = append(seriesPaths, child)
			pen.Series = append(pen.Series, scanSeries(child, e.Name()))
			continue
		}
		if looksBook(child) {
			pen.Books = append(pen.Books, treeBook(child, e.Name()))
			if hasSiblingSeries(entries, path) {
				pen.Problems = append(pen.Problems, TreeProblem{
					Path:    child,
					Message: fmt.Sprintf("%q looks like a book next to a series. Put series books under Books/.", e.Name()),
				})
			}
			continue
		}
		pen.Problems = append(pen.Problems, TreeProblem{
			Path:    child,
			Message: fmt.Sprintf("%q is not a series or book (need 00_Series-Bible, Books, or 01_Manuscript).", e.Name()),
		})
	}
	_ = seriesPaths
	sort.Slice(pen.Series, func(i, j int) bool {
		return strings.ToLower(pen.Series[i].Name) < strings.ToLower(pen.Series[j].Name)
	})
	sort.Slice(pen.Books, func(i, j int) bool {
		return strings.ToLower(pen.Books[i].Folder) < strings.ToLower(pen.Books[j].Folder)
	})
	return pen
}

func scanSeries(path, name string) TreeSeries {
	se := TreeSeries{Path: path, Name: name, Books: []TreeBook{}, Problems: []TreeProblem{}}
	booksDir := filepath.Join(path, "Books")
	info, err := os.Stat(booksDir)
	if err != nil || !info.IsDir() {
		se.Problems = append(se.Problems, TreeProblem{Path: path, Message: "Series has no Books folder."})
		return se
	}
	entries, err := os.ReadDir(booksDir)
	if err != nil {
		se.Problems = append(se.Problems, TreeProblem{Path: booksDir, Message: err.Error()})
		return se
	}
	for _, e := range entries {
		if !e.IsDir() || skipSyncDir(e.Name()) {
			continue
		}
		child := filepath.Join(booksDir, e.Name())
		if looksBook(child) || looksLooseBook(child) {
			se.Books = append(se.Books, treeBook(child, e.Name()))
			continue
		}
		se.Problems = append(se.Problems, TreeProblem{
			Path:    child,
			Message: fmt.Sprintf("%q in Books/ does not look like a book (need 01_Manuscript or 02_Story-Elements).", e.Name()),
		})
	}
	sort.Slice(se.Books, func(i, j int) bool {
		return strings.ToLower(se.Books[i].Folder) < strings.ToLower(se.Books[j].Folder)
	})
	return se
}

func treeBook(path, folder string) TreeBook {
	return TreeBook{Path: path, Folder: folder, Name: displayBookName(folder)}
}

func displayBookName(folder string) string {
	return bookFolderPrefix.ReplaceAllString(folder, "")
}

func looksSeries(path string) bool {
	return hasDir(path, "00_Series-Bible") || hasDir(path, "Books")
}

func looksBook(path string) bool {
	return hasDir(path, "01_Manuscript") || hasDir(path, "02_Story-Elements")
}

func looksLooseBook(path string) bool {
	return looksBook(path)
}

func hasSiblingSeries(entries []os.DirEntry, penPath string) bool {
	for _, e := range entries {
		if e.IsDir() && looksSeries(filepath.Join(penPath, e.Name())) {
			return true
		}
	}
	return false
}

func hasDir(parent, name string) bool {
	info, err := os.Stat(filepath.Join(parent, name))
	return err == nil && info.IsDir()
}

func CreateSeriesOnDisk(writingRoot, penName, penPath, seriesName string) (string, error) {
	parent, err := resolvePenDir(writingRoot, penName, penPath)
	if err != nil {
		return "", err
	}
	return ApplySeriesTemplate(parent, seriesName)
}

func CreateBookOnDisk(writingRoot, penName, penPath, seriesPath, bookTitle string) (string, error) {
	if seriesPath != "" {
		parent := filepath.Join(seriesPath, "Books")
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return "", err
		}
		return ApplyBookTemplate(parent, bookTitle, filepath.Base(seriesPath))
	}
	parent, err := resolvePenDir(writingRoot, penName, penPath)
	if err != nil {
		return "", err
	}
	return ApplyBookTemplate(parent, bookTitle, "")
}

func resolvePenDir(writingRoot, penName, penPath string) (string, error) {
	if info, err := os.Stat(penPath); err == nil && info.IsDir() {
		return penPath, nil
	}
	return EnsurePenDir(writingRoot, penName)
}

func RenameProjectDir(path, newDisplay string) (string, error) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("folder not found")
	}
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	prefix := bookFolderPrefix.FindString(base)
	next := filepath.Join(dir, prefix+safeFolderName(newDisplay))
	if next == path {
		return path, nil
	}
	if err := os.Rename(path, next); err != nil {
		return "", err
	}
	return next, nil
}

func DeleteProjectDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("missing path")
	}
	return os.RemoveAll(path)
}
