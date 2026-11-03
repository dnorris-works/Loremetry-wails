package store

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"strings"
)

// ReportContentOnly strips the standard About/How to use header from a formatted report body.
func ReportContentOnly(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	if i := strings.Index(body, "\n---\n"); i >= 0 {
		return strings.TrimSpace(body[i+5:])
	}
	if strings.HasPrefix(body, "---\n") {
		return strings.TrimSpace(body[4:])
	}
	return strings.TrimSpace(body)
}

// MarkdownToDocx builds a minimal .docx from simple markdown (# headings and paragraphs).
func MarkdownToDocx(markdown string) ([]byte, error) {
	md := strings.ReplaceAll(markdown, "\r\n", "\n")
	md = strings.TrimSpace(md)
	var body strings.Builder
	body.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	body.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	body.WriteString(`<w:body>`)

	blocks := strings.Split(md, "\n\n")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		lines := strings.Split(block, "\n")
		first := strings.TrimSpace(lines[0])
		if strings.HasPrefix(first, "# ") {
			writeDocxHeading(&body, 1, strings.TrimSpace(first[2:]))
			rest := strings.TrimSpace(strings.Join(lines[1:], "\n"))
			if rest != "" {
				writeDocxParagraph(&body, rest)
			}
			continue
		}
		if strings.HasPrefix(first, "## ") {
			writeDocxHeading(&body, 2, strings.TrimSpace(first[3:]))
			rest := strings.TrimSpace(strings.Join(lines[1:], "\n"))
			if rest != "" {
				writeDocxParagraph(&body, rest)
			}
			continue
		}
		if strings.HasPrefix(first, "### ") {
			writeDocxHeading(&body, 3, strings.TrimSpace(first[4:]))
			rest := strings.TrimSpace(strings.Join(lines[1:], "\n"))
			if rest != "" {
				writeDocxParagraph(&body, rest)
			}
			continue
		}
		writeDocxParagraph(&body, strings.Join(lines, "\n"))
	}

	body.WriteString(`<w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr>`)
	body.WriteString(`</w:body></w:document>`)

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`,
		"word/_rels/document.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`,
		"word/document.xml": body.String(),
	}
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			_ = zw.Close()
			return nil, err
		}
		if _, err := w.Write([]byte(content)); err != nil {
			_ = zw.Close()
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeDocxHeading(b *strings.Builder, level int, text string) {
	style := "Heading1"
	if level == 2 {
		style = "Heading2"
	} else if level >= 3 {
		style = "Heading3"
	}
	fmt.Fprintf(b, `<w:p><w:pPr><w:pStyle w:val="%s"/></w:pPr><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, style, xmlText(text))
}

func writeDocxParagraph(b *strings.Builder, text string) {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	fmt.Fprintf(b, `<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlText(text))
}

func xmlText(s string) string {
	return html.EscapeString(s)
}
