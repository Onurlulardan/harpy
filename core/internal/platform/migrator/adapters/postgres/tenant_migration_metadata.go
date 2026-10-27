package postgres

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

type tenantMigrationMetadata struct {
	owner        string
	version      int64
	origin       postgresv1.DatabaseOrigin
	selectedPath string
	selectedHash string
	runID        string
}

func (metadata tenantMigrationMetadata) identity(
	target TenantMigrationTarget,
	attempt int,
) TenantCheckpointIdentity {
	return TenantCheckpointIdentity{
		TenantID: target.TenantID, Owner: metadata.owner,
		Version: metadata.version, Origin: string(metadata.origin),
		SelectedPath: metadata.selectedPath, SelectedHash: metadata.selectedHash,
		RunID: metadata.runID, Attempt: attempt,
	}
}

func (metadata tenantMigrationMetadata) sentinelOrigin() string {
	if metadata.origin == postgresv1.DatabaseOriginV1Adopted {
		return "v1_adopted"
	}
	return "new"
}
