package cutover

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

const migrationSourceDirectory = "core/internal/platform/cutover/migrations"

func selectMigrationSupersessions(
	documents []migration.Document,
) ([]migration.Document, error) {
	catalog, err := migrationsupersession.DecodeCatalog(migrationSupersessionCatalog)
	if err != nil {
		return nil, fmt.Errorf("decode platform cutover migration supersessions: %w", err)
	}
	if catalog.Owner() != "platform_cutover" || catalog.Scope() != "platform" ||
		catalog.SourceDirectory() != migrationSourceDirectory {
		return nil, fmt.Errorf("platform cutover migration supersession identity differs")
	}
	indexed := make(map[string]migration.Document, len(documents))
	for _, document := range documents {
		indexed[document.Path()] = document
	}
	superseded := make(map[string]bool, len(catalog.Entries()))
	replacements := make(map[string]migrationsupersession.Entry, len(catalog.Entries()))
	for _, entry := range catalog.Entries() {
		if !strings.HasSuffix(entry.OldPath(), ".fresh.sql") {
			continue
		}
		oldDocument, oldExists := indexed[entry.OldPath()]
		replacement, replacementExists := indexed[entry.ReplacementPath()]
		if !oldExists || !replacementExists ||
			migrationDocumentSHA256(oldDocument.Content()) != entry.OldSHA256() ||
			migrationDocumentSHA256(replacement.Content()) != entry.ReplacementSHA256() {
			return nil, fmt.Errorf(
				"select platform cutover migration supersessions: migration supersession path or SHA-256 differs",
			)
		}
		superseded[entry.OldPath()] = true
		replacements[entry.ReplacementPath()] = entry
	}
	return bindFreshMigrationDocuments(documents, superseded, replacements)
}

func bindFreshMigrationDocuments(
	documents []migration.Document,
	superseded map[string]bool,
	replacements map[string]migrationsupersession.Entry,
) ([]migration.Document, error) {
	reservedPrefix := migrationSourceDirectory + "/supersessions/"
	selected := make([]migration.Document, 0, len(documents))
	for _, document := range documents {
		if superseded[document.Path()] {
			continue
		}
		bound, err := bindFreshMigrationDocument(document, replacements, reservedPrefix)
		if err != nil {
			return nil, err
		}
		selected = append(selected, bound)
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	return selected, nil
}

func bindFreshMigrationDocument(
	document migration.Document,
	replacements map[string]migrationsupersession.Entry,
	reservedPrefix string,
) (migration.Document, error) {
	if entry, replaces := replacements[document.Path()]; replaces {
		bound, err := migration.NewSupersedingDocument(
			"platform_cutover", "platform", document.Path(), document.Content(), entry,
		)
		if err != nil {
			return migration.Document{}, fmt.Errorf(
				"bind platform cutover migration supersession: %w", err,
			)
		}
		return bound, nil
	}
	if strings.HasPrefix(document.Path(), reservedPrefix) {
		return migration.Document{}, fmt.Errorf(
			"select platform cutover migration supersessions: migration supersession replacement is orphaned",
		)
	}
	bound, err := migration.NewDocument(
		"platform_cutover", "platform", document.Path(), document.Content(),
	)
	if err != nil {
		return migration.Document{}, fmt.Errorf(
			"bind platform cutover migration supersession: %w", err,
		)
	}
	return bound, nil
}

func migrationDocumentSHA256(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}
