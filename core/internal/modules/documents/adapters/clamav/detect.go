package clamav

import (
	"bytes"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

var ErrFileSignature = errors.New("documents file signature mismatches declaration")

func detect(filename, declaredMIME string, prefix []byte) (domain.FileKind, string, error) {
	expected, err := domain.DeclaredFileKind(filename, declaredMIME, 1)
	if err != nil || len(prefix) == 0 || !matchesSignature(expected, prefix) {
		return "", "", ErrFileSignature
	}
	return expected, declaredMIME, nil
}

func matchesSignature(kind domain.FileKind, value []byte) bool {
	switch kind {
	case domain.FilePDF:
		return bytes.HasPrefix(value, []byte("%PDF-"))
	case domain.FileJPEG:
		return has(value, 0xff, 0xd8, 0xff)
	case domain.FilePNG:
		return bytes.HasPrefix(value, []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	case domain.FileGIF:
		return bytes.HasPrefix(value, []byte("GIF87a")) || bytes.HasPrefix(value, []byte("GIF89a"))
	case domain.FileWebP:
		return len(value) >= 12 && string(value[:4]) == "RIFF" && string(value[8:12]) == "WEBP"
	case domain.FileDOC, domain.FileXLS, domain.FilePPT:
		return bytes.HasPrefix(value, []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1})
	case domain.FileDOCX:
		return zipMarker(value, "word/")
	case domain.FileXLSX:
		return zipMarker(value, "xl/")
	case domain.FilePPTX:
		return zipMarker(value, "ppt/")
	case domain.FileODT:
		return zipMarker(value, "application/vnd.oasis.opendocument.text")
	case domain.FileODS:
		return zipMarker(value, "application/vnd.oasis.opendocument.spreadsheet")
	case domain.FileODP:
		return zipMarker(value, "application/vnd.oasis.opendocument.presentation")
	case domain.FileJSON:
		trimmed := bytes.TrimSpace(value)
		return validText(value) && len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[')
	case domain.FileTXT, domain.FileCSV:
		return validText(value)
	default:
		return false
	}
}

func zipMarker(value []byte, marker string) bool {
	zip := bytes.HasPrefix(value, []byte{'P', 'K', 3, 4}) ||
		bytes.HasPrefix(value, []byte{'P', 'K', 5, 6})
	return zip && bytes.Contains(value, []byte(marker))
}

func validText(value []byte) bool {
	return utf8.Valid(value) && !bytes.ContainsRune(value, 0) &&
		strings.TrimSpace(string(value)) != ""
}

func has(value []byte, signature ...byte) bool {
	return bytes.HasPrefix(value, signature)
}
