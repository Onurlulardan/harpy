package domain

import (
	"errors"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const MaxUploadBytes int64 = 100 * 1024 * 1024

var ErrFilePolicy = errors.New("documents file policy rejects input")

type FileKind string

const (
	FilePDF  FileKind = "pdf"
	FileTXT  FileKind = "txt"
	FileCSV  FileKind = "csv"
	FileJSON FileKind = "json"
	FileJPEG FileKind = "jpeg"
	FilePNG  FileKind = "png"
	FileGIF  FileKind = "gif"
	FileWebP FileKind = "webp"
	FileDOC  FileKind = "doc"
	FileDOCX FileKind = "docx"
	FileXLS  FileKind = "xls"
	FileXLSX FileKind = "xlsx"
	FilePPT  FileKind = "ppt"
	FilePPTX FileKind = "pptx"
	FileODT  FileKind = "odt"
	FileODS  FileKind = "ods"
	FileODP  FileKind = "odp"
)

type FileDeclaration struct {
	Filename     string
	DeclaredMIME string
	DetectedKind FileKind
	Size         int64
}

type fileContract struct {
	kind FileKind
	mime string
}

var fileContracts = map[string][]fileContract{
	".pdf":  {{FilePDF, "application/pdf"}},
	".txt":  {{FileTXT, "text/plain"}},
	".csv":  {{FileCSV, "text/csv"}},
	".json": {{FileJSON, "application/json"}},
	".jpg":  {{FileJPEG, "image/jpeg"}}, ".jpeg": {{FileJPEG, "image/jpeg"}},
	".png": {{FilePNG, "image/png"}}, ".gif": {{FileGIF, "image/gif"}},
	".webp": {{FileWebP, "image/webp"}},
	".doc":  {{FileDOC, "application/msword"}},
	".docx": {{FileDOCX, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}},
	".xls":  {{FileXLS, "application/vnd.ms-excel"}},
	".xlsx": {{FileXLSX, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}},
	".ppt":  {{FilePPT, "application/vnd.ms-powerpoint"}},
	".pptx": {{FilePPTX, "application/vnd.openxmlformats-officedocument.presentationml.presentation"}},
	".odt":  {{FileODT, "application/vnd.oasis.opendocument.text"}},
	".ods":  {{FileODS, "application/vnd.oasis.opendocument.spreadsheet"}},
	".odp":  {{FileODP, "application/vnd.oasis.opendocument.presentation"}},
}

func ValidateFile(declaration FileDeclaration) error {
	if declaration.Size < 1 || declaration.Size > MaxUploadBytes ||
		!validFilename(declaration.Filename) {
		return ErrFilePolicy
	}
	extension := strings.ToLower(filepath.Ext(declaration.Filename))
	for _, contract := range fileContracts[extension] {
		if declaration.DetectedKind == contract.kind && declaration.DeclaredMIME == contract.mime {
			return nil
		}
	}
	return ErrFilePolicy
}

func DeclaredFileKind(filename, declaredMIME string, size int64) (FileKind, error) {
	if size < 1 || size > MaxUploadBytes || !validFilename(filename) {
		return "", ErrFilePolicy
	}
	for _, contract := range fileContracts[strings.ToLower(filepath.Ext(filename))] {
		if declaredMIME == contract.mime {
			return contract.kind, nil
		}
	}
	return "", ErrFilePolicy
}

func validFilename(value string) bool {
	if value == "" || len(value) > 255 || !utf8.ValidString(value) ||
		strings.ContainsAny(value, "/\\") || value == "." || value == ".." {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}
