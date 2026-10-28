package migrationmanifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

func applySourceSupersessions(
	repositoryRoot string,
	spec SourceSpec,
	documents []SourceDocument,
) ([]SourceDocument, error) {
	catalogPath, info, found, err := resolveSupersessionCatalog(
		repositoryRoot, spec.Directory,
	)
	if err != nil {
		return nil, err
	}
	if !found {
		if rejectErr := migrationsupersession.RejectReservedPaths(
			spec.Directory, supersessionDocuments(documents),
		); rejectErr != nil {
			return nil, rejectErr
		}
		return documents, nil
	}
	absoluteCatalogPath := filepath.Join(repositoryRoot, filepath.FromSlash(catalogPath))
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("migration supersession catalog is not a regular file")
	}
	encoded, err := os.ReadFile(absoluteCatalogPath)
	if err != nil {
		return nil, fmt.Errorf("read migration supersession catalog: %w", err)
	}
	catalog, err := migrationsupersession.DecodeCatalog(encoded)
	if err != nil {
		return nil, err
	}
	if catalog.Owner() != spec.Owner || catalog.Scope() != spec.Scope ||
		catalog.SourceDirectory() != spec.Directory {
		return nil, fmt.Errorf("migration supersession catalog differs from source claim")
	}
	selected, err := migrationsupersession.Select(
		catalog, supersessionDocuments(documents),
	)
	if err != nil {
		return nil, err
	}
	result := make([]SourceDocument, len(selected))
	for index, document := range selected {
		result[index], err = NewSourceDocument(spec.Owner, document.Path, document.Content)
		if err != nil {
			return nil, fmt.Errorf("bind migration supersession source: %w", err)
		}
	}
	return result, nil
}

func supersessionDocuments(documents []SourceDocument) []migrationsupersession.Source {
	result := make([]migrationsupersession.Source, len(documents))
	for index, document := range documents {
		result[index] = migrationsupersession.Source{
			Path: document.Path(), Content: document.Content(),
		}
	}
	return result
}
