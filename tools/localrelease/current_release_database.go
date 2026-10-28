package main

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

const currentReleaseAppliedSQL = `SELECT pg_catalog.count(*)=1
FROM platform.schema_contracts AS contract
JOIN platform.release_generations AS active_release
  ON active_release.migration_run_id=contract.verified_run_id
 AND active_release.database_generation=contract.database_generation
 AND active_release.minimum_runtime_generation=contract.minimum_runtime_generation
 AND active_release.migration_manifest_sha256=contract.migration_manifest_sha256
 AND active_release.extension_manifest_sha256=contract.extension_manifest_sha256
 AND active_release.postgres_image_digest=$6
JOIN platform.environment_identity AS identity
  ON identity.singleton=TRUE
 AND identity.environment_id=active_release.environment_id
 AND identity.database_origin=active_release.database_origin
JOIN platform.migration_runs AS run
  ON run.run_id=active_release.migration_run_id
 AND run.release_id=active_release.release_id
 AND run.environment_id=active_release.environment_id
 AND run.database_origin=active_release.database_origin
 AND run.command='up' AND run.state='succeeded'
 AND run.migration_manifest_sha256=active_release.migration_manifest_sha256
 AND run.extension_manifest_sha256=active_release.extension_manifest_sha256
JOIN platform.bootstrap_receipts AS bootstrap
  ON bootstrap.kind=active_release.bootstrap_receipt_kind
 AND bootstrap.run_id=active_release.bootstrap_receipt_run_id
 AND bootstrap.release_id=active_release.release_id
 AND bootstrap.environment_id=active_release.environment_id
 AND bootstrap.database_origin=active_release.database_origin
 AND bootstrap.database_system_identifier=identity.database_system_identifier
 AND bootstrap.database_oid=identity.database_oid
JOIN platform.bootstrap_receipts AS extension
  ON extension.kind=active_release.extension_receipt_kind
 AND extension.run_id=active_release.extension_receipt_run_id
 AND extension.release_id=active_release.release_id
 AND extension.environment_id=active_release.environment_id
 AND extension.database_origin=active_release.database_origin
 AND extension.database_system_identifier=identity.database_system_identifier
 AND extension.database_oid=identity.database_oid
JOIN platform.bootstrap_receipts AS cleanup
  ON cleanup.kind=active_release.cleanup_receipt_kind
 AND cleanup.run_id=active_release.cleanup_receipt_run_id
 AND cleanup.release_id=active_release.release_id
 AND cleanup.environment_id=active_release.environment_id
 AND cleanup.database_origin=active_release.database_origin
 AND cleanup.database_system_identifier=identity.database_system_identifier
 AND cleanup.database_oid=identity.database_oid
WHERE contract.component='core' AND contract.status='ready'
  AND active_release.environment_id=$1 AND active_release.database_origin='fresh'
  AND contract.database_generation >= $3
  AND $2 >= contract.minimum_runtime_generation
  AND EXISTS (
SELECT 1 FROM platform.release_generations AS compatible_release
WHERE compatible_release.environment_id=active_release.environment_id
  AND compatible_release.database_origin=active_release.database_origin
  AND compatible_release.database_generation <= contract.database_generation
  AND compatible_release.runtime_generation=$2
  AND compatible_release.minimum_database_generation=$3
  AND compatible_release.migration_manifest_sha256=$4
  AND compatible_release.extension_manifest_sha256=$5
  AND compatible_release.postgres_image_digest=$6)`

func validateCurrentReleaseContract(contract runtimev1.BinarySchemaContract) error {
	if err := contract.Validate(); err != nil {
		return fmt.Errorf("current release contract is invalid: %w", err)
	}
	if contract.MigrationManifestSHA256 == foundationMigrationDigest {
		return fmt.Errorf("current release status rejects the pinned bootstrap image")
	}
	return nil
}

func verifyCurrentReleaseApplied(
	ctx context.Context,
	target *pgx.Conn,
	environmentID string,
	contract runtimev1.BinarySchemaContract,
) error {
	var applied bool
	err := target.QueryRow(ctx, currentReleaseAppliedSQL,
		environmentID, contract.RuntimeGeneration, contract.MinimumDatabaseGeneration,
		contract.MigrationManifestSHA256, contract.ExtensionManifestSHA256,
		contract.PostgresImageDigest,
	).Scan(&applied)
	if err != nil {
		return fmt.Errorf("verify current release evidence: %w", err)
	}
	if !applied {
		return fmt.Errorf("current embedded release is not finalized and admissible")
	}
	return nil
}

func runReleaseApplied(ctx context.Context, arguments []string) error {
	path, err := parseSinglePath(arguments, "release-applied", "admin-dsn-file")
	if err != nil {
		return err
	}
	contract, err := readEmbeddedBuildContract()
	if err != nil {
		return err
	}
	if err := validateCurrentReleaseContract(contract); err != nil {
		return err
	}
	dsn, err := readSecret(path, "admin DSN")
	if err != nil {
		return err
	}
	operationContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	admin, config, err := openAdmin(operationContext, dsn)
	if err != nil {
		return err
	}
	defer admin.Close(context.Background())
	target, err := openTarget(operationContext, config)
	if err != nil {
		return err
	}
	defer target.Close(context.Background())
	live, err := readLiveDatabaseIdentity(operationContext, target)
	if err != nil {
		return err
	}
	environmentID, err := readExistingEnvironmentID(operationContext, target, live)
	if err != nil {
		return err
	}
	if environmentID == "" {
		return fmt.Errorf("current release environment is absent")
	}
	if err := verifyCurrentReleaseApplied(
		operationContext, target, environmentID, contract,
	); err != nil {
		return err
	}
	if err := verifyCurrentSignedRelease(time.Now().UTC()); err != nil {
		return err
	}
	fmt.Println("current embedded release already finalized")
	return nil
}
