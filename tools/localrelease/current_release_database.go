package main

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

const currentReleaseAppliedSQL = `SELECT pg_catalog.count(*)=1
FROM platform.environment_identity AS identity
WHERE identity.singleton=TRUE
  AND identity.environment_id=$1::uuid
  AND identity.database_origin='fresh'
  AND EXISTS (
    SELECT 1 FROM platform.goose_db_version AS history
    WHERE history.is_applied AND history.version_id=20261103000004
  )`

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
) error {
	var applied bool
	err := target.QueryRow(ctx, currentReleaseAppliedSQL, environmentID).Scan(&applied)
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
		operationContext, target, environmentID,
	); err != nil {
		return err
	}
	if err := verifyCurrentSignedRelease(time.Now().UTC()); err != nil {
		return err
	}
	fmt.Println("current embedded release already finalized")
	return nil
}
