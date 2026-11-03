package postgres

import (
	"context"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

const baselineFoundationSQL = `SELECT identity.database_origin,
identity.database_system_identifier, identity.database_oid,
identity.environment_id::text, identity.release_id::text,
extension.extversion, namespace.nspname,
pg_catalog.pg_get_userbyid(extension.extowner),
(SELECT count(*) FROM platform.goose_db_version AS history
 WHERE history.is_applied AND history.version_id IN (20261028000000, 20261102000000))
FROM platform.environment_identity AS identity
JOIN pg_catalog.pg_extension AS extension ON extension.extname='vector'
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=extension.extnamespace
WHERE identity.singleton=TRUE`

func readBaselineFoundation(
	ctx context.Context,
	queryer roleQueryer,
	binary runtimev1.BinarySchemaContract,
) (foundationPreflight, error) {
	var origin, systemID, environmentID, releaseID string
	var extensionVersion, extensionSchema, extensionOwner string
	var databaseOID, appliedVersions int64
	err := queryer.QueryRow(ctx, baselineFoundationSQL).Scan(
		&origin, &systemID, &databaseOID, &environmentID, &releaseID,
		&extensionVersion, &extensionSchema, &extensionOwner, &appliedVersions,
	)
	if err != nil {
		return foundationPreflight{}, fmt.Errorf("%w: baseline foundation query: %w", ErrStartupPreflight, err)
	}
	return assembleBaselineFoundation(
		origin, systemID, environmentID, releaseID,
		extensionVersion, extensionSchema, extensionOwner,
		databaseOID, appliedVersions, binary,
	)
}

func assembleBaselineFoundation(
	origin, systemID, environmentID, releaseID string,
	extensionVersion, extensionSchema, extensionOwner string,
	databaseOID, appliedVersions int64,
	binary runtimev1.BinarySchemaContract,
) (foundationPreflight, error) {
	if appliedVersions != 2 {
		return foundationPreflight{}, fmt.Errorf("%w: baseline foundation is incomplete", ErrStartupPreflight)
	}
	return foundationPreflight{
		count: 1, origin: origin, systemIdentifier: systemID, databaseOID: databaseOID,
		environmentID: environmentID, bootstrapReleaseID: releaseID, activeReleaseID: releaseID,
		runID: releaseID, runState: "succeeded", runCommand: "up",
		runMigrationSHA256: binary.MigrationManifestSHA256,
		runExtensionSHA256: binary.ExtensionManifestSHA256,
		databaseGeneration: binary.MinimumDatabaseGeneration, minimumRuntimeGeneration: 1,
		contractMigrationSHA256: binary.MigrationManifestSHA256,
		contractExtensionSHA256: binary.ExtensionManifestSHA256,
		contractStatus:          "ready", contractVerifiedRunID: releaseID, contractRevision: 1,
		extensionVersion: extensionVersion, extensionSchema: extensionSchema,
		extensionOwner: extensionOwner, extensionOrigin: origin,
		extensionVerifiedRunID: releaseID, binaryReleaseKnown: true,
	}, nil
}
