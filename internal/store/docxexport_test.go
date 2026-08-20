package store

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestReportContentOnly(t *testing.T) {
	in := "# Title\n\n## About this report\n\nAbout.\n\n## How to use\n\nUse.\n\n---\n\n# Ch1\n\nHello.\n"
	got := ReportContentOnly(in)
	if !strings.Contains(got, "# Ch1") || strings.Contains(got, "About this report") {
		t.Fatalf("%q", got)
	}
}

func TestMarkdownToDocx(t *testing.T) {
	raw, err := MarkdownToDocx("# Chapter One\n\nHello world.\n\n## Scene\n\nMore text.")
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 100 {
		t.Fatalf("docx too small: %d", len(raw))
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			found = true
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(rc)
			_ = rc.Close()
			xml := buf.String()
			if !strings.Contains(xml, "Chapter One") || !strings.Contains(xml, "Hello world") {
				t.Fatalf("missing text: %s", xml)
			}
		}
	}
	if !found {
		t.Fatal("missing document.xml")
	}
}
