package migrator

import (
	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func tenantCheckpointIdentity(
	target postgresstore.TenantMigrationTarget,
	selected postgresv1.SelectedMigration,
	runID string,
	attempt int,
) postgresstore.TenantCheckpointIdentity {
	return postgresstore.TenantCheckpointIdentity{
		TenantID: target.TenantID, Owner: selected.Owner, Version: selected.Version,
		Origin: string(selected.Origin), SelectedPath: selected.Path,
		SelectedHash: selected.SHA256, RunID: runID, Attempt: attempt,
	}
}

func tenantSentinelOrigin(origin postgresv1.DatabaseOrigin) string {
	if origin == postgresv1.DatabaseOriginV1Adopted {
		return "v1_adopted"
	}
	return "new"
}
