package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	xmlTagRe      = regexp.MustCompile(`<[^>]*>`)
	paragraphEndRe = regexp.MustCompile(`(?i)</w:p>`)
	insBlockRe     = regexp.MustCompile(`(?s)<w:ins[^>]*>(.*?)</w:ins>`)
	commentBlockRe = regexp.MustCompile(`(?s)<w:comment\s[^>]*w:author="([^"]+)"[^>]*>(.*?)</w:comment>`)
)

func ReadDocxText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var documentXML *zip.File
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			documentXML = f
			break
		}
	}

	if documentXML == nil {
		return "", fmt.Errorf("word/document.xml not found in docx")
	}

	rc, err := documentXML.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, rc); err != nil {
		return "", err
	}

	content := buf.String()
	content = paragraphEndRe.ReplaceAllString(content, "\n")
	content = xmlTagRe.ReplaceAllString(content, "")

	return content, nil
}

func ExtractDocxTrackChanges(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var insertions []string
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			break
		}
		var buf bytes.Buffer
		io.Copy(&buf, rc)
		rc.Close()

		content := buf.String()
		matches := insBlockRe.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			text := strings.TrimSpace(xmlTagRe.ReplaceAllString(m[1], ""))
			if text != "" {
				insertions = append(insertions, "+ "+text)
			}
		}
		break
	}

	var comments []string
	for _, f := range r.File {
		if f.Name != "word/comments.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			break
		}
		var buf bytes.Buffer
		io.Copy(&buf, rc)
		rc.Close()

		content := buf.String()
		matches := commentBlockRe.FindAllStringSubmatch(content, -1)
		for i, m := range matches {
			author := m[1]
			text := strings.TrimSpace(xmlTagRe.ReplaceAllString(m[2], ""))
			if text != "" {
				comments = append(comments, fmt.Sprintf("Comment #%d (%s): %s", i+1, author, text))
			}
		}
		break
	}

	var parts []string
	if len(insertions) > 0 {
		parts = append(parts, "=== INSERTED TEXT (Track Changes) ===")
		parts = append(parts, strings.Join(insertions, "\n"))
	}
	if len(comments) > 0 {
		parts = append(parts, "=== LECTURER COMMENTS ===")
		parts = append(parts, strings.Join(comments, "\n"))
	}
	if len(parts) == 0 {
		return "(No track changes or comments found in this document)", nil
	}
	return strings.Join(parts, "\n\n"), nil
}
