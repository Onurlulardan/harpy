package clamav

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestDetectRejectsExtensionMagicMismatch(t *testing.T) {
	if _, _, err := detect("report.pdf", "application/pdf",
		[]byte("<html>not a PDF</html>")); err == nil {
		t.Fatal("mismatched PDF signature was accepted")
	}
}

func TestDetectDistinguishesOOXMLFamilies(t *testing.T) {
	value := append([]byte{'P', 'K', 3, 4}, []byte("[Content_Types].xml word/document.xml")...)
	kind, _, err := detect("report.docx",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document", value)
	if err != nil || kind != domain.FileDOCX {
		t.Fatalf("kind=%s err=%v", kind, err)
	}
	if _, _, err = detect("report.xlsx",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", value); err == nil {
		t.Fatal("DOCX bytes were accepted as XLSX")
	}
}

func TestParseVersionExposesDatabaseAge(t *testing.T) {
	revision := "ClamAV 1.4.2/27890/Sat Jul 18 09:34:31 2026"
	parsed, at, err := parseVersion(revision)
	if err != nil || parsed != revision || at.Year() != 2026 || at.Day() != 18 {
		t.Fatalf("revision=%q at=%v err=%v", parsed, at, err)
	}
}
