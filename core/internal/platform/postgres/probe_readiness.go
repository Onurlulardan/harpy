package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (pool *Pool) ProbeReadiness(ctx context.Context) ReadinessObservation {
	result := ReadinessObservation{schemaReason: SchemaMissing}
	if pool == nil || pool.closed.Load() || pool.probeConfig == nil || ctx == nil {
		return result
	}
	connection, err := pgx.ConnectConfig(ctx, pool.probeConfig.Copy())
	if err != nil {
		return result
	}
	defer closeProbeConnection(connection, pool.cleanupTimeout)
	transaction, err := connection.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return result
	}
	defer rollbackProbeTransaction(transaction, pool.cleanupTimeout)
	catalog, err := readCatalogPreflight(ctx, transaction)
	if err != nil || !catalogMatchesStartup(catalog, pool.startup) {
		return result
	}
	result.database = true
	if _, err := transaction.Exec(ctx, pool.probeRoleSQL()); err != nil {
		return result
	}
	if err := validateRolePreflight(ctx, transaction, pool.contract); err != nil {
		return result
	}
	if err := validateSessionState(ctx, transaction, pool.contract); err != nil {
		return result
	}
	result.databaseRole = true
	return result
}

func (pool *Pool) probeRoleSQL() string {
	switch {
	case pool.profile == ProfileAPI || foundationAPIProbe(pool.profile):
		return probeAPIRoleSQL
	case pool.profile == ProfileWorker:
		return probeRuntimeRoleSQL
	default:
		return probeOwnerRoleSQL
	}
}

func foundationAPIProbe(profile Profile) bool {
	switch profile {
	case ProfileOperatorInitialize, ProfileOperatorIdentityReconcile,
		ProfileOperatorLicenseReconcile, ProfileDatabaseBackupExecutor:
		return true
	default:
		return false
	}
}
