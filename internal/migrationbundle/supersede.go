package migrationbundle

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

func dropSuperseded(set corpusSet, documents []Source, catalogPath string) ([]Source, error) {
	encoded, err := os.ReadFile(catalogPath)
	if err != nil {
		return nil, err
	}
	catalog, err := migrationsupersession.DecodeCatalog(encoded)
	if err != nil {
		return nil, err
	}
	indexed := make(map[string]Source, len(documents))
	for _, document := range documents {
		indexed[document.Path] = document
	}
	superseded := map[string]bool{}
	replacements := map[string]bool{}
	for _, entry := range catalog.Entries() {
		if err = acceptSupersession(indexed, entry, superseded, replacements); err != nil {
			return nil, err
		}
	}
	reserved := set.root + "/migrations/supersessions/"
	selected := make([]Source, 0, len(documents))
	for _, document := range documents {
		if superseded[document.Path] {
			continue
		}
		if strings.HasPrefix(document.Path, reserved) && !replacements[document.Path] {
			return nil, fmt.Errorf("fresh migration supersession %s is orphaned", document.Path)
		}
		selected = append(selected, document)
	}
	return selected, nil
}

func acceptSupersession(
	indexed map[string]Source, entry migrationsupersession.Entry,
	superseded, replacements map[string]bool,
) error {
	if !strings.HasSuffix(entry.OldPath(), ".fresh.sql") {
		return nil
	}
	oldDocument, oldExists := indexed[entry.OldPath()]
	replacement, replacementExists := indexed[entry.ReplacementPath()]
	if !oldExists || !replacementExists ||
		sourceDigest(oldDocument.Body) != entry.OldSHA256() ||
		sourceDigest(replacement.Body) != entry.ReplacementSHA256() {
		return fmt.Errorf("fresh migration supersession %s differs", entry.OldPath())
	}
	superseded[entry.OldPath()] = true
	replacements[entry.ReplacementPath()] = true
	return nil
}

func sourceDigest(body []byte) string {
	sum := sha256.Sum256(body)
	return fmt.Sprintf("%x", sum)
}
