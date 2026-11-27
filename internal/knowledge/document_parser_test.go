// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package knowledge

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, name, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// writeDocx builds a minimal .docx with the given word/document.xml body.
func writeDocx(t *testing.T, bodyXML string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "design.docx")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	doc := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` +
		bodyXML + `</w:body></w:document>`
	if _, err := w.Write([]byte(doc)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNormalizeDocument_TextFormatsPassThrough(t *testing.T) {
	for _, ext := range []string{".md", ".markdown", ".txt", ".rst", ".MD"} {
		body := "## Decision\n\n- Payment service MUST use idempotency keys\n"
		got, err := NormalizeDocument(writeFile(t, "adr"+ext, body))
		if err != nil {
			t.Fatalf("%s: %v", ext, err)
		}
		if got != body {
			t.Fatalf("%s: content changed: %q", ext, got)
		}
	}
}

func TestNormalizeDocument_PDFIsRefusedClearly(t *testing.T) {
	_, err := NormalizeDocument(writeFile(t, "spec.pdf", "%PDF-1.7 binary"))
	if !errors.Is(err, ErrPDFNotSupported) {
		t.Fatalf("expected ErrPDFNotSupported, got %v", err)
	}
	if !strings.Contains(err.Error(), "spec.pdf") {
		t.Fatalf("error should name the file: %v", err)
	}
}

func TestNormalizeDocument_UnsupportedExtension(t *testing.T) {
	if _, err := NormalizeDocument(writeFile(t, "notes.odt", "x")); err == nil {
		t.Fatal("expected an error for .odt")
	}
}

func TestNormalizeDocument_MissingFile(t *testing.T) {
	if _, err := NormalizeDocument(filepath.Join(t.TempDir(), "missing.md")); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestNormalizeDocument_DocxHeadingsListsAndText(t *testing.T) {
	body := `` +
		`<w:p><w:pPr><w:pStyle w:val="Title"/></w:pPr><w:r><w:t>ADR 7: Refunds</w:t></w:r></w:p>` +
		`<w:p><w:pPr><w:pStyle w:val="Heading2"/></w:pPr><w:r><w:t>Decision</w:t></w:r></w:p>` +
		`<w:p><w:pPr><w:numPr><w:ilvl w:val="0"/></w:numPr></w:pPr>` +
		`<w:r><w:t xml:space="preserve">Payment service MUST </w:t></w:r><w:r><w:t>use idempotency keys</w:t></w:r></w:p>` +
		`<w:p><w:r><w:t>Plain</w:t><w:tab/><w:t>paragraph</w:t><w:br/><w:t>continues</w:t></w:r></w:p>` +
		`<w:p></w:p>` +
		`<w:p><w:r><w:t>Fish &amp; chips</w:t></w:r></w:p>`
	got, err := NormalizeDocument(writeDocx(t, body))
	if err != nil {
		t.Fatalf("NormalizeDocument: %v", err)
	}
	want := "# ADR 7: Refunds\n" +
		"## Decision\n" +
		"- Payment service MUST use idempotency keys\n" +
		"Plain\tparagraph continues\n" +
		"Fish & chips\n"
	if got != want {
		t.Fatalf("docx text =\n%q\nwant\n%q", got, want)
	}
}

func TestNormalizeDocument_DocxWithoutDocumentXML(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.docx")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	if _, err := zw.Create("word/styles.xml"); err != nil {
		t.Fatal(err)
	}
	zw.Close()
	f.Close()
	if _, err := NormalizeDocument(p); err == nil || !strings.Contains(err.Error(), "word/document.xml") {
		t.Fatalf("expected a clear error about the missing document part, got %v", err)
	}
}

func TestNormalizeDocument_NotAZip(t *testing.T) {
	if _, err := NormalizeDocument(writeFile(t, "fake.docx", "plain text, not a zip")); err == nil {
		t.Fatal("expected an error for a .docx that is not a zip archive")
	}
}

func TestHeadingLevel(t *testing.T) {
	cases := map[string]int{
		"Title": 1, "Heading1": 1, "heading 3": 3, "Heading6": 6,
		"Heading7": 0, "Heading10": 0, "Normal": 0, "ListParagraph": 0, "": 0,
	}
	for style, want := range cases {
		if got := headingLevel(style); got != want {
			t.Errorf("headingLevel(%q) = %d, want %d", style, got, want)
		}
	}
}
