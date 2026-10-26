package postgres

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/google/uuid"
)

func NewTenantStore(
	runID string,
	target TenantMigrationTarget,
	session TenantSessionProof,
	selected []postgresv1.SelectedMigration,
) (*TenantStore, error) {
	parsedTenant, err := uuid.Parse(target.TenantID)
	if err != nil || parsedTenant == uuid.Nil || parsedTenant.String() != target.TenantID ||
		target.SchemaName == "" || session.BackendPID <= 0 ||
		session.SchemaName != target.SchemaName {
		return nil, fmt.Errorf("tenant migration target is invalid")
	}
	canonical, err := canonicalSelectedMigrations(selected)
	if err != nil {
		return nil, fmt.Errorf("validate tenant migration metadata: %w", err)
	}
	metadata := make(map[int64]tenantMigrationMetadata, len(canonical))
	for _, candidate := range canonical {
		if candidate.Scope != "tenant" || candidate.TransactionMode != "transactional" {
			return nil, fmt.Errorf("tenant migration selection is invalid")
		}
		metadata[candidate.Version] = tenantMigrationMetadata{
			owner: candidate.Owner, version: candidate.Version, origin: candidate.Origin,
			selectedPath: candidate.Path, selectedHash: candidate.SHA256, runID: runID,
		}
	}
	return &TenantStore{
		target: target, session: session, metadataByVersion: metadata,
		finalVersion: canonical[len(canonical)-1].Version,
	}, nil
}
