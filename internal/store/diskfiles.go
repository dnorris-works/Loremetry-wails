package store

import (
	"bytes"
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

type FileSection struct {
	Index int    `json:"index"`
	Title string `json:"title"`
}

type FileSections struct {
	Dir      string        `json:"dir"`
	Name     string        `json:"name"`
	Sections []FileSection `json:"sections"`
	Total    int           `json:"total"`
}

type FileSectionContent struct {
	Dir     string `json:"dir"`
	Name    string `json:"name"`
	Index   int    `json:"index"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Total   int    `json:"total"`
}

const chunkSizeThreshold = 50_000
const maxSectionBytes = 30_000

func isHeading(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "# ") {
		return strings.TrimSpace(trimmed[2:]), true
	}
	if strings.HasPrefix(trimmed, "## ") {
		return strings.TrimSpace(trimmed[3:]), true
	}
	return "", false
}

type rawSection struct {
	title string
	lines []string
	size  int
}

func splitRaw(text string) []rawSection {
	lines := strings.Split(text, "\n")
	var out []rawSection
	cur := rawSection{title: "Top"}
	for _, line := range lines {
		if title, ok := isHeading(line); ok {
			out = append(out, cur)
			cur = rawSection{title: title}
		}
		cur.lines = append(cur.lines, line)
		cur.size += len(line) + 1
	}
	out = append(out, cur)

	// If any section exceeds maxSectionBytes, subdivide it by paragraph breaks.
	var result []rawSection
	for _, sec := range out {
		if sec.size <= maxSectionBytes {
			result = append(result, sec)
			continue
		}
		part := rawSection{title: sec.title}
		partNum := 1
		for _, line := range sec.lines {
			part.lines = append(part.lines, line)
			part.size += len(line) + 1
			if part.size >= maxSectionBytes && strings.TrimSpace(line) == "" {
				result = append(result, part)
				partNum++
				part = rawSection{title: fmt.Sprintf("%s (cont. %d)", sec.title, partNum)}
			}
		}
		if len(part.lines) > 0 {
			result = append(result, part)
		}
	}
	return result
}

func splitSections(text string) []FileSection {
	raw := splitRaw(text)
	secs := make([]FileSection, len(raw))
	for i, r := range raw {
		secs[i] = FileSection{Index: i, Title: r.title}
	}
	return secs
}

func splitSectionTexts(text string) []string {
	raw := splitRaw(text)
	out := make([]string, len(raw))
	for i, r := range raw {
		out[i] = strings.Join(r.lines, "\n")
	}
	return out
}

func readAndSplit(dir, name string) (string, []rawSection, error) {
	path, err := diskFilePath(dir, name)
	if err != nil {
		return "", nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return filepath.Base(path), nil, err
	}
	return filepath.Base(path), splitRaw(string(b)), nil
}

func ReadDiskFileSections(dir, name string) (FileSections, error) {
	base, raw, err := readAndSplit(dir, name)
	if err != nil {
		return FileSections{}, err
	}
	secs := make([]FileSection, len(raw))
	for i, r := range raw {
		secs[i] = FileSection{Index: i, Title: r.title}
	}
	return FileSections{Dir: dir, Name: base, Sections: secs, Total: len(secs)}, nil
}

func ReadDiskFileSection(dir, name string, index int) (FileSectionContent, error) {
	base, raw, err := readAndSplit(dir, name)
	if err != nil {
		return FileSectionContent{}, err
	}
	if index < 0 || index >= len(raw) {
		return FileSectionContent{}, fmt.Errorf("section index out of range")
	}
	return FileSectionContent{
		Dir:   dir,
		Name:  base,
		Index: index,
		Title: raw[index].title,
		Text:  strings.Join(raw[index].lines, "\n"),
		Total: len(raw),
	}, nil
}

// OpenDiskFile returns the full file for small files, or metadata + first page for large files.
type OpenedFile struct {
	Dir      string `json:"dir"`
	Name     string `json:"name"`
	Text     string `json:"text,omitempty"`
	Large    bool   `json:"large"`
	FileSize int64  `json:"file_size,omitempty"`
}

func OpenDiskFile(dir, name string) (OpenedFile, error) {
	path, err := diskFilePath(dir, name)
	if err != nil {
		return OpenedFile{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return OpenedFile{}, err
	}
	base := filepath.Base(path)
	if info.Size() <= int64(chunkSizeThreshold) {
		b, err := os.ReadFile(path)
		if err != nil {
			return OpenedFile{}, err
		}
		return OpenedFile{Dir: dir, Name: base, Text: string(b)}, nil
	}
	return OpenedFile{Dir: dir, Name: base, Large: true, FileSize: info.Size()}, nil
}

// ReadDiskFileRange reads a byte range from a file, snapping to line boundaries.
type FileRange struct {
	Dir      string `json:"dir"`
	Name     string `json:"name"`
	Text     string `json:"text"`
	Start    int64  `json:"start"`
	End      int64  `json:"end"`
	FileSize int64  `json:"file_size"`
}

const defaultPageSize = 32_000

func ReadDiskFileRange(dir, name string, offset int64, limit int) (FileRange, error) {
	path, err := diskFilePath(dir, name)
	if err != nil {
		return FileRange{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return FileRange{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return FileRange{}, err
	}
	size := info.Size()
	if offset < 0 {
		offset = 0
	}
	if offset >= size {
		return FileRange{Dir: dir, Name: filepath.Base(path), Start: size, End: size, FileSize: size}, nil
	}
	if limit <= 0 {
		limit = defaultPageSize
	}

	// Snap offset backward to a line boundary (unless at 0).
	if offset > 0 {
		if _, err := f.Seek(offset, 0); err != nil {
			return FileRange{}, err
		}
		// Read backward up to 512 bytes to find the previous newline.
		back := int64(512)
		if back > offset {
			back = offset
		}
		buf := make([]byte, back)
		if _, err := f.ReadAt(buf, offset-back); err != nil {
			return FileRange{}, err
		}
		idx := bytes.LastIndexByte(buf, '\n')
		if idx >= 0 {
			offset = offset - back + int64(idx) + 1
		} else {
			offset = offset - back
		}
	}

	end := offset + int64(limit)
	if end > size {
		end = size
	}
	// Snap end forward to a line boundary.
	if end < size {
		if _, err := f.Seek(end, 0); err != nil {
			return FileRange{}, err
		}
		trail := make([]byte, 512)
		n, _ := f.Read(trail)
		if n > 0 {
			idx := bytes.IndexByte(trail[:n], '\n')
			if idx >= 0 {
				end += int64(idx) + 1
			} else {
				end += int64(n)
			}
		}
	}

	buf := make([]byte, end-offset)
	if _, err := f.ReadAt(buf, offset); err != nil && err.Error() != "EOF" {
		return FileRange{}, err
	}
	return FileRange{
		Dir:      dir,
		Name:     filepath.Base(path),
		Text:     string(buf),
		Start:    offset,
		End:      end,
		FileSize: size,
	}, nil
}

func WriteDiskFileSection(dir, name string, index int, sectionText string) (DiskFile, error) {
	path, err := diskFilePath(dir, name)
	if err != nil {
		return DiskFile{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return DiskFile{}, err
	}
	chunks := splitSectionTexts(string(b))
	if index < 0 || index >= len(chunks) {
		return DiskFile{}, fmt.Errorf("section index out of range")
	}
	chunks[index] = sectionText
	full := strings.Join(chunks, "\n")
	if err := writeTextFile(path, full); err != nil {
		return DiskFile{}, err
	}
	return DiskFile{Dir: dir, Name: filepath.Base(path), Path: path}, nil
}

func NeedsChunking(dir, name string) bool {
	path, err := diskFilePath(dir, name)
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Size() > chunkSizeThreshold
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
	if base == "01_Chapters" || base == "00_Current-Draft" {
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
