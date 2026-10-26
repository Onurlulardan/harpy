package migration

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

type expectedDocument struct {
	owner           string
	scope           string
	transactionMode string
	sha256          string
}

func bindDocuments(
	manifest postgresv1.MigrationManifest,
	origin postgresv1.DatabaseOrigin,
	selected []postgresv1.SelectedMigration,
	documents []Document,
) (map[string][]byte, error) {
	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("validate migration manifest: %w", err)
	}
	expected := expectedDocuments(manifest)
	provided := make(map[string]Document, len(documents))
	for _, document := range documents {
		if _, duplicate := provided[document.path]; duplicate {
			return nil, fmt.Errorf("migration source %s is duplicated", document.path)
		}
		provided[document.path] = document
	}
	if len(provided) != len(expected) {
		return nil, fmt.Errorf("embedded migration set differs from manifest")
	}
	for path, contract := range expected {
		document, exists := provided[path]
		if !exists {
			return nil, fmt.Errorf("manifest migration %s is not embedded", path)
		}
		if err := validateDocument(document, contract); err != nil {
			return nil, fmt.Errorf("migration %s: %w", path, err)
		}
	}
	return selectDocumentBytes(origin, selected, provided)
}

func expectedDocuments(manifest postgresv1.MigrationManifest) map[string]expectedDocument {
	expected := make(map[string]expectedDocument, len(manifest.Entries)*2)
	for _, entry := range manifest.Entries {
		for _, variant := range []*postgresv1.MigrationVariant{
			entry.Variants.Fresh, entry.Variants.V1Adopted,
		} {
			if variant != nil {
				expected[variant.Path] = expectedDocument{
					owner: entry.Owner, scope: entry.Scope,
					transactionMode: entry.TransactionMode,
					sha256:          variant.SHA256,
				}
			}
		}
	}
	return expected
}

func selectDocumentBytes(
	origin postgresv1.DatabaseOrigin,
	selected []postgresv1.SelectedMigration,
	provided map[string]Document,
) (map[string][]byte, error) {
	if !origin.Valid() || len(selected) == 0 {
		return nil, fmt.Errorf("selected migration origin or entries are invalid")
	}
	bound := make(map[string][]byte, len(selected))
	for _, migration := range selected {
		if migration.Origin != origin {
			return nil, fmt.Errorf("selected migration origin differs")
		}
		document, exists := provided[migration.Path]
		if !exists {
			return nil, fmt.Errorf("selected migration %s is not embedded", migration.Path)
		}
		bound[migration.Path] = append([]byte(nil), document.content...)
	}
	return bound, nil
}
