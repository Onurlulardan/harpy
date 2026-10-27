package migrationsupersession

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

const maximumSourceBytes = 1024 * 1024

type Source struct {
	Path    string
	Content []byte
}

func Select(catalog Catalog, documents []Source) ([]Source, error) {
	if err := validateCatalog(catalog); err != nil {
		return nil, err
	}
	indexed, err := indexDocuments(catalog.sourceDirectory, documents)
	if err != nil {
		return nil, err
	}
	replacements := make(map[string]bool, len(catalog.entries))
	superseded := make(map[string]bool, len(catalog.entries))
	for _, entry := range catalog.entries {
		oldDocument, oldExists := indexed[entry.oldPath]
		replacement, replacementExists := indexed[entry.replacementPath]
		if !oldExists || !replacementExists ||
			documentSHA256(oldDocument.Content) != entry.oldSHA256 ||
			documentSHA256(replacement.Content) != entry.replacementSHA256 {
			return nil, fmt.Errorf("migration supersession path or SHA-256 differs")
		}
		superseded[entry.oldPath] = true
		replacements[entry.replacementPath] = true
	}
	selected := make([]Source, 0, len(documents)-len(superseded))
	reservedPrefix := catalog.sourceDirectory + "/supersessions/"
	for _, document := range documents {
		if strings.HasPrefix(document.Path, reservedPrefix) && !replacements[document.Path] {
			return nil, fmt.Errorf("migration supersession replacement is orphaned")
		}
		if !superseded[document.Path] {
			selected = append(selected, cloneSource(document))
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path < selected[right].Path
	})
	return selected, nil
}

func RejectReservedPaths(sourceDirectory string, documents []Source) error {
	if !validSourceDirectory(sourceDirectory) {
		return fmt.Errorf(
			"migration supersession source directory %q is invalid", sourceDirectory,
		)
	}
	prefix := sourceDirectory + "/supersessions/"
	for _, document := range documents {
		if strings.HasPrefix(document.Path, prefix) {
			return fmt.Errorf("migration supersession catalog is missing")
		}
	}
	return nil
}

func indexDocuments(sourceDirectory string, documents []Source) (map[string]Source, error) {
	if len(documents) == 0 || len(documents) > maximumSourceDocuments {
		return nil, fmt.Errorf("migration supersession source set size is invalid")
	}
	prefix := sourceDirectory + "/"
	indexed := make(map[string]Source, len(documents))
	for _, document := range documents {
		if !fs.ValidPath(document.Path) || !strings.HasPrefix(document.Path, prefix) ||
			len(document.Content) == 0 || len(document.Content) > maximumSourceBytes {
			return nil, fmt.Errorf("migration supersession source document is invalid")
		}
		if _, exists := indexed[document.Path]; exists {
			return nil, fmt.Errorf("migration supersession source path is duplicate")
		}
		indexed[document.Path] = cloneSource(document)
	}
	return indexed, nil
}

func documentSHA256(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}

func cloneSource(document Source) Source {
	return Source{Path: document.Path, Content: append([]byte(nil), document.Content...)}
}
