package migrator

import (
	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func tenantCheckpointMatchesVerifiedSelection(
	checkpoint postgresstore.TenantCheckpoint,
	expected postgresstore.TenantCheckpointIdentity,
	selected postgresv1.SelectedMigration,
	aliases map[int64]historySupersession,
) bool {
	if postgresstore.TenantCheckpointMatchesSelection(checkpoint, expected) {
		return true
	}
	alias, exists := aliases[selected.Version]
	return exists && alias.matchesTenantCheckpoint(selected, checkpoint, expected)
}

func (value historySupersession) matchesTenantCheckpoint(
	selected postgresv1.SelectedMigration,
	checkpoint postgresstore.TenantCheckpoint,
	expected postgresstore.TenantCheckpointIdentity,
) bool {
	if !value.matchesCurrent(selected) || value.scope != tenantMigrationScope ||
		checkpoint.TenantID != expected.TenantID ||
		checkpoint.Owner != expected.Owner ||
		checkpoint.Version != expected.Version ||
		checkpoint.Origin != string(value.origin) {
		return false
	}
	for _, identity := range value.superseded {
		if checkpoint.SelectedPath == identity.path &&
			checkpoint.SelectedHash == identity.sha256 {
			return true
		}
	}
	return false
}
