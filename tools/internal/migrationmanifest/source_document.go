package migrationmanifest

import (
	"bytes"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
)

const maximumDocumentBytes = 1024 * 1024

var (
	sourceOwnerPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	dollarQuotePattern = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)?\$`)
)

type SourceDocument struct {
	owner   string
	path    string
	content []byte
}

func NewSourceDocument(owner, path string, content []byte) (SourceDocument, error) {
	if !sourceOwnerPattern.MatchString(owner) {
		return SourceDocument{}, fmt.Errorf("migration source owner is invalid")
	}
	if !fs.ValidPath(path) || !strings.HasSuffix(path, ".sql") {
		return SourceDocument{}, fmt.Errorf("migration source path is invalid")
	}
	if len(content) == 0 || len(content) > maximumDocumentBytes {
		return SourceDocument{}, fmt.Errorf("migration source size is invalid")
	}
	if err := validateGooseStatementGroups(content); err != nil {
		return SourceDocument{}, err
	}
	return SourceDocument{
		owner: owner, path: path, content: append([]byte(nil), content...),
	}, nil
}

func validateGooseStatementGroups(content []byte) error {
	grouped := false
	for _, line := range bytes.Split(content, []byte("\n")) {
		switch string(bytes.TrimSpace(line)) {
		case "-- +goose StatementBegin":
			if grouped {
				return fmt.Errorf("migration source has nested goose statement groups")
			}
			grouped = true
		case "-- +goose StatementEnd":
			if !grouped {
				return fmt.Errorf("migration source has unmatched goose statement end")
			}
			grouped = false
		default:
			if dollarQuotePattern.Match(line) && !grouped {
				return fmt.Errorf("dollar-quoted SQL requires a goose statement group")
			}
		}
	}
	if grouped {
		return fmt.Errorf("migration source has an unclosed goose statement group")
	}
	return nil
}

func (document SourceDocument) Owner() string {
	return document.owner
}

func (document SourceDocument) Path() string {
	return document.path
}

func (document SourceDocument) Content() []byte {
	return append([]byte(nil), document.content...)
}
