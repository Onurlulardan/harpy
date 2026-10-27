package migration

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func (registry *ReleaseRegistry) SelectReleaseDocuments(
	artifact attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest],
	documents []Document,
) ([]Document, error) {
	if err := registry.Authorize(artifact); err != nil {
		return nil, err
	}
	return registry.selectAuthorizedDocuments(artifact.Value(), documents)
}

func (registry *ReleaseRegistry) selectAuthorizedDocuments(
	manifest postgresv1.MigrationManifest,
	documents []Document,
) ([]Document, error) {
	if registry == nil || registry.documentContracts == nil {
		return nil, fmt.Errorf("migration release document registry is unavailable")
	}
	target := expectedDocuments(manifest)
	selected := make([]Document, 0, len(target))
	seen := make(map[string]bool, len(documents))
	for _, document := range documents {
		if seen[document.path] {
			return nil, fmt.Errorf("embedded migration source is duplicated")
		}
		seen[document.path] = true
		metadata, err := InspectDocument(document)
		if err != nil {
			return nil, err
		}
		contract := expectedDocument{
			owner: document.owner, scope: document.scope,
			transactionMode: metadata.TransactionMode, sha256: metadata.SHA256,
		}
		if _, authorized := registry.documentContracts[document.path][contract]; !authorized {
			return nil, fmt.Errorf("embedded migration source is not release-authorized")
		}
		if _, required := target[document.path]; required {
			selected = append(selected, document)
		}
	}
	if len(selected) != len(target) {
		return nil, fmt.Errorf("release migration source set is incomplete")
	}
	return selected, nil
}
