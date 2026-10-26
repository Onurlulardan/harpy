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
	result.schemaReason = SchemaContractMalformed
	foundation, err := readFoundationPreflight(ctx, transaction, pool.binaryContract)
	if err != nil {
		return result
	}
	origin, err := validateFoundationEvidence(foundation, catalog)
	if err != nil || origin != pool.startup.databaseOrigin {
		return result
	}
	result.schemaReason = foundationSchemaReadiness(foundation, pool.binaryContract)
	result.schemaContract = result.schemaReason == SchemaReady
	return result
}

func (pool *Pool) probeRoleSQL() string {
	if pool.contract.runtime {
		return probeRuntimeRoleSQL
	}
	if pool.profile == ProfileDatabaseBackupExecutor {
		return probeBackupRoleSQL
	}
	return probeOwnerRoleSQL
}
