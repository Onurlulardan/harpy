package migration

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func NewTenantCatalog(
	registry *ReleaseRegistry,
	artifact attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest],
	origin postgresv1.DatabaseOrigin,
	documents []Document,
) (*Catalog, error) {
	if err := registry.Authorize(artifact); err != nil {
		return nil, err
	}
	selected, err := artifact.Value().Select(origin)
	if err != nil {
		return nil, fmt.Errorf("select tenant migration release: %w", err)
	}
	bound, err := bindDocuments(artifact.Value(), origin, selected, documents)
	if err != nil {
		return nil, err
	}
	files, scopedSelected, err := partitionCatalog(bound, selected)
	if err != nil {
		return nil, err
	}
	supersessions, err := bindHistorySupersessions(selected, documents)
	if err != nil {
		return nil, err
	}
	return &Catalog{selected: scopedSelected, files: files,
		supersessions: supersessions}, nil
}
