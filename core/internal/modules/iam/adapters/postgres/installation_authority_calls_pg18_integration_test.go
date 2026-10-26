//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

var fixedIAMContractTime = time.Date(
	2026, time.July, 29, 12, 0, 0, 0, time.UTC,
)

func callLegacyInstallationOwnerPG18(
	ctx context.Context, database iamPlatformGuardDatabase, operationID uuid.UUID,
) error {
	connection, err := database.runtime.Conn(ctx)
	if err != nil {
		return err
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := transaction.ExecContext(ctx, `
SET LOCAL application_name TO 'harpy-v2:core-operator-initialize'`); err != nil {
		_ = transaction.Rollback()
		return err
	}
	var replay bool
	err = transaction.QueryRowContext(ctx, `
SELECT platform.iam_installation_instance_owner_bind_v1(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,1,$10)`,
		operationID, installationPG18EnvironmentID, installationPG18TenantID,
		installationPG18Schema, installationPG18UserID, uuid.New(), uuid.New(),
		uuid.New(), uuid.New(), installationPG18Manifest).Scan(&replay)
	if err != nil {
		_ = transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

func callIAMTenantInitializePG18(
	ctx context.Context, database iamPlatformGuardDatabase,
	routine string, operationID uuid.UUID,
) (bool, error) {
	statement := `SELECT ` + routine + `(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`
	var replay bool
	err := database.runtime.QueryRowContext(ctx, statement,
		operationID, installationPG18EnvironmentID, installationPG18TenantID,
		uuid.New(), uuid.New(), uuid.New(), "admin@harpy.io", "Admin",
		"Harpy", installationPG18Manifest, uuid.New(), uuid.New(),
		uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		uuid.New(), uuid.New(), uuid.New(), uuid.New()).Scan(&replay)
	return replay, err
}

func assertInstallationRetainedACLPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	var runtimeAllowed, loginAllowed bool
	err := database.migrator.QueryRowContext(ctx, `
SELECT
 pg_catalog.has_function_privilege(
  'harpy_runtime',
  'platform.iam_installation_instance_owner_input_v1(uuid)','EXECUTE'),
 pg_catalog.has_function_privilege(
  'harpy_v2_runtime',
  'platform.iam_installation_instance_owner_input_v1(uuid)','EXECUTE')`).
		Scan(&runtimeAllowed, &loginAllowed)
	if err != nil {
		t.Fatal(err)
	}
	if !runtimeAllowed || loginAllowed {
		t.Fatalf("retained input ACL runtime/login = %t/%t",
			runtimeAllowed, loginAllowed)
	}
	var rows int
	if err := database.runtime.QueryRowContext(ctx, `
SELECT count(*) FROM platform.iam_installation_instance_owner_input_v1($1)`,
		installationPG18EnvironmentID).Scan(&rows); err != nil {
		t.Fatalf("retained input call: %v", err)
	}
	var internal bool
	err = database.runtime.QueryRowContext(ctx, `
SELECT platform.tenancy_activation_step_consume_v1($1,$2)`,
		uuid.New(), "iam.initialize_tenant_identity").Scan(&internal)
	assertIAMSQLState(t, err, "42501")
}

func assertInstallationSuccessorReplayPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	epochOperationID := uuid.MustParse("50000000-0000-4000-8000-000000000001")
	epochEnvironmentID := uuid.MustParse("50000000-0000-4000-8000-000000000002")
	seedInstallationCapabilityPG18(t, ctx, database, epochOperationID,
		"iam.initialize_tenant_identity", "environment_initialize")
	if _, err := database.migrator.ExecContext(ctx, `
UPDATE platform.test_activation_capabilities SET environment_id=$2
WHERE operation_id=$1`, epochOperationID, epochEnvironmentID); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		var initialized bool
		err := database.runtime.QueryRowContext(ctx, `
SELECT platform.iam_installation_security_epoch_initialize_v2(
    $1,$2,$3,$4,$5)`,
			epochOperationID, epochEnvironmentID, installationPG18TenantID,
			installationPG18Schema, fixedIAMContractTime).Scan(&initialized)
		if err != nil || !initialized {
			t.Fatalf("epoch successor attempt %d initialized=%t: %v",
				attempt+1, initialized, err)
		}
	}
	var epochRows int
	if err := database.migrator.QueryRowContext(ctx, `
SELECT count(*) FROM platform.iam_security_epochs WHERE environment_id=$1`,
		epochEnvironmentID).Scan(&epochRows); err != nil {
		t.Fatal(err)
	}
	if epochRows != 1 {
		t.Fatalf("epoch successor replay rows = %d", epochRows)
	}
	tenantOperationID := uuid.MustParse("50000000-0000-4000-8000-000000000003")
	seedInstallationCapabilityPG18(t, ctx, database, tenantOperationID,
		"iam.initialize_tenant_identity", "environment_initialize")
	if _, err := database.migrator.ExecContext(ctx, `
UPDATE platform.test_activation_capabilities SET schema_name='pg_catalog'
WHERE operation_id=$1`, tenantOperationID); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := callIAMTenantInitializePG18(
			ctx, database, "iam_installation_initialize_v2", tenantOperationID,
		); err != nil {
			t.Fatalf("tenant successor attempt %d: %v", attempt+1, err)
		}
	}
	var tenantRows int
	if err := database.migrator.QueryRowContext(ctx, `
SELECT count(*) FROM iam_capability.test_identity_writes WHERE operation_id=$1`,
		tenantOperationID).Scan(&tenantRows); err != nil {
		t.Fatal(err)
	}
	if tenantRows != 1 {
		t.Fatalf("tenant successor replay rows = %d", tenantRows)
	}
}
