package migration

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

const (
	maximumDocumentBytes     = 1024 * 1024
	maximumSupersessionDepth = 3
)

var ownerPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
var documentSHA256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Document struct {
	owner      string
	scope      string
	path       string
	content    []byte
	superseded []documentIdentity
}

type documentIdentity struct {
	path   string
	sha256 string
}

func NewSupersedingDocument(
	owner, scope, path string,
	content []byte,
	entry migrationsupersession.Entry,
) (Document, error) {
	return NewSupersedingDocumentChain(
		owner, scope, path, content, []migrationsupersession.Entry{entry},
	)
}

func NewSupersedingDocumentChain(
	owner, scope, path string,
	content []byte,
	lineage []migrationsupersession.Entry,
) (Document, error) {
	document, err := NewDocument(owner, scope, path, content)
	if err != nil {
		return Document{}, err
	}
	contentHash := fmt.Sprintf("%x", sha256.Sum256(content))
	if len(lineage) == 0 || len(lineage) > maximumSupersessionDepth {
		return Document{}, fmt.Errorf("migration supersession document binding is invalid")
	}
	seen := make(map[string]bool, len(lineage))
	for index, entry := range lineage {
		if entry.OldPath() == path || seen[entry.OldPath()] ||
			!fs.ValidPath(entry.OldPath()) ||
			!strings.HasSuffix(entry.OldPath(), ".sql") ||
			!documentSHA256Pattern.MatchString(entry.OldSHA256()) {
			return Document{}, fmt.Errorf("migration supersession document binding is invalid")
		}
		if index > 0 {
			previous := lineage[index-1]
			if previous.ReplacementPath() != entry.OldPath() ||
				previous.ReplacementSHA256() != entry.OldSHA256() {
				return Document{}, fmt.Errorf("migration supersession document chain differs")
			}
		}
		seen[entry.OldPath()] = true
		document.superseded = append(document.superseded, documentIdentity{
			path: entry.OldPath(), sha256: entry.OldSHA256(),
		})
	}
	terminal := lineage[len(lineage)-1]
	if terminal.ReplacementPath() != path ||
		terminal.ReplacementSHA256() != contentHash {
		return Document{}, fmt.Errorf("migration supersession document binding is invalid")
	}
	return document, nil
}

func NewDocument(owner, scope, path string, content []byte) (Document, error) {
	if !ownerPattern.MatchString(owner) {
		return Document{}, fmt.Errorf("migration document owner is invalid")
	}
	if scope != "platform" && scope != "tenant" {
		return Document{}, fmt.Errorf("migration document scope is invalid")
	}
	if !fs.ValidPath(path) || !strings.HasSuffix(path, ".sql") {
		return Document{}, fmt.Errorf("migration document path is invalid")
	}
	if len(content) == 0 || len(content) > maximumDocumentBytes {
		return Document{}, fmt.Errorf("migration document size is invalid")
	}
	return Document{
		owner: owner, scope: scope, path: path, content: append([]byte(nil), content...),
	}, nil
}

func (document Document) Scope() string {
	return document.scope
}

func (document Document) Owner() string {
	return document.owner
}

func (document Document) Path() string {
	return document.path
}

func (document Document) Content() []byte {
	return append([]byte(nil), document.content...)
}
