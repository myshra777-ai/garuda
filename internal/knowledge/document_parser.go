// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package knowledge

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// maxDocxXMLBytes caps how much of word/document.xml is read, so a
// malformed or hostile file cannot exhaust memory.
const maxDocxXMLBytes = 32 << 20

// ErrPDFNotSupported is returned for PDF files. Extracting text from a PDF
// reliably needs a real PDF parser; feeding raw PDF bytes to the claim
// parser only produced noise, so PDFs are refused with a clear message.
var ErrPDFNotSupported = errors.New("PDF text extraction is not supported yet: export the PDF to Markdown, DOCX or plain text and ingest that file")

// NormalizeDocument converts a supported document into Markdown-shaped text
// that the ADR parser can read.
//
//   - .md, .markdown: returned as is
//   - .txt, .rst: returned as plain text
//   - .docx: paragraphs extracted from word/document.xml; Word headings
//     become Markdown headings and list items become "- " bullets, so the
//     parser can find decision sections and bullet rules
//   - .pdf: refused with ErrPDFNotSupported
func NormalizeDocument(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".md", ".markdown", ".txt", ".rst":
		content, err := os.ReadFile(filePath)
		return string(content), err
	case ".pdf":
		return "", fmt.Errorf("%s: %w", filepath.Base(filePath), ErrPDFNotSupported)
	case ".docx":
		return parseDocxText(filePath)
	default:
		return "", fmt.Errorf("unsupported document format: %s", ext)
	}
}

// parseDocxText extracts paragraph text from a DOCX package (a zip archive
// containing word/document.xml), using only the standard library.
func parseDocxText(filePath string) (string, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("open docx: %w", err)
	}
	defer zr.Close()

	var doc *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			doc = f
			break
		}
	}
	if doc == nil {
		return "", fmt.Errorf("open docx: word/document.xml not found in %s", filepath.Base(filePath))
	}

	rc, err := doc.Open()
	if err != nil {
		return "", fmt.Errorf("open docx: %w", err)
	}
	defer rc.Close()

	return docxXMLToMarkdown(io.LimitReader(rc, maxDocxXMLBytes))
}

// docxXMLToMarkdown walks WordprocessingML and emits one line per paragraph.
func docxXMLToMarkdown(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)

	var (
		out        strings.Builder
		para       strings.Builder
		inPara     bool
		inText     bool
		headingLvl int
		isListItem bool
	)

	flush := func() {
		text := strings.TrimSpace(para.String())
		if text != "" {
			switch {
			case headingLvl > 0:
				out.WriteString(strings.Repeat("#", headingLvl) + " ")
			case isListItem:
				out.WriteString("- ")
			}
			out.WriteString(text)
			out.WriteString("\n")
		}
		para.Reset()
		headingLvl = 0
		isListItem = false
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("parse docx: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				if inPara {
					flush()
				}
				inPara = true
			case "pStyle":
				if lvl := headingLevel(attrValue(t, "val")); lvl > 0 {
					headingLvl = lvl
				}
			case "numPr":
				isListItem = true
			case "t":
				inText = true
			case "tab":
				para.WriteString("\t")
			case "br", "cr":
				para.WriteString(" ")
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				flush()
				inPara = false
			}
		case xml.CharData:
			if inText {
				para.Write(t)
			}
		}
	}
	if inPara {
		flush()
	}
	return out.String(), nil
}

// headingLevel maps Word paragraph style IDs to Markdown heading levels:
// "Title" -> 1, "Heading1".."Heading6" -> 1..6. Anything else is body text.
func headingLevel(style string) int {
	s := strings.ToLower(strings.ReplaceAll(style, " ", ""))
	if s == "title" {
		return 1
	}
	if strings.HasPrefix(s, "heading") && len(s) == len("heading")+1 {
		if d := s[len(s)-1]; d >= '1' && d <= '6' {
			return int(d - '0')
		}
	}
	return 0
}

func attrValue(el xml.StartElement, local string) string {
	for _, a := range el.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}
