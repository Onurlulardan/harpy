//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

var (
	legacyOwnerOperationID   = uuid.MustParse("40000000-0000-4000-8000-000000000001")
	legacyTenantOperationID  = uuid.MustParse("40000000-0000-4000-8000-000000000002")
	legacyEpochEnvironmentID = uuid.MustParse("40000000-0000-4000-8000-000000000003")
)

func assertInstallationLegacyAuthorityPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	if err := callLegacyInstallationOwnerPG18(ctx, database, legacyOwnerOperationID); err != nil {
		t.Fatalf("Gen70 legacy owner call: %v", err)
	}
	var initialized bool
	err := database.runtime.QueryRowContext(ctx, `
SELECT platform.iam_security_epoch_initialize_v1($1,$2)`,
		legacyEpochEnvironmentID, fixedIAMContractTime).Scan(&initialized)
	if err != nil || !initialized {
		t.Fatalf("Gen70 legacy epoch call initialized=%t: %v", initialized, err)
	}
	if _, err := callIAMTenantInitializePG18(
		ctx, database, "iam_installation_initialize_v1", legacyTenantOperationID,
	); err != nil {
		t.Fatalf("Gen70 legacy tenant call: %v", err)
	}
	assertInstallationLegacyWriteCountsPG18(t, ctx, database)
}

func applyInstallationAuthorityContractsPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	origin string,
) {
	t.Helper()
	platformPath := filepath.Join(
		"..", "..", "migrations", "platform", origin,
		"20260729100038_contract_iam_installation_platform_authority_v1."+
			origin+".sql",
	)
	platformBody, err := os.ReadFile(platformPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.migrator.ExecContext(ctx, string(platformBody)); err != nil {
		t.Fatalf("apply IAM platform authority contract: %v", err)
	}
	connection, err := database.migrator.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, `
SELECT pg_catalog.set_config(
    'search_path','iam_capability,pg_catalog',false
)`); err != nil {
		t.Fatal(err)
	}
	tenantPath := filepath.Join(
		"..", "..", "migrations", "tenant", origin,
		"20260729100046_contract_iam_installation_tenant_authority_v1."+
			origin+".sql",
	)
	tenantBody, err := os.ReadFile(tenantPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(ctx, string(tenantBody)); err != nil {
		t.Fatalf("apply IAM tenant authority contract: %v", err)
	}
}

func assertInstallationContractedAuthorityPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	assertIAMSQLState(t,
		callLegacyInstallationOwnerPG18(ctx, database, uuid.New()), "42501")
	var initialized bool
	err := database.runtime.QueryRowContext(ctx, `
SELECT platform.iam_security_epoch_initialize_v1($1,$2)`,
		uuid.New(), fixedIAMContractTime).Scan(&initialized)
	assertIAMSQLState(t, err, "42501")
	_, err = callIAMTenantInitializePG18(
		ctx, database, "iam_installation_initialize_v1", uuid.New(),
	)
	assertIAMSQLState(t, err, "42501")
	assertInstallationLegacyWriteCountsPG18(t, ctx, database)
	assertInstallationRetainedACLPG18(t, ctx, database)
	assertInstallationSuccessorReplayPG18(t, ctx, database)
}

func assertInstallationLegacyWriteCountsPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	var ownerWrites, epochWrites, tenantWrites int
	err := database.migrator.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.test_owner_writes WHERE operation_id=$1),
 (SELECT count(*) FROM platform.test_legacy_epoch_writes WHERE environment_id=$2),
 (SELECT count(*) FROM iam_capability.test_identity_writes WHERE operation_id=$3)`,
		legacyOwnerOperationID, legacyEpochEnvironmentID, legacyTenantOperationID).
		Scan(&ownerWrites, &epochWrites, &tenantWrites)
	if err != nil {
		t.Fatal(err)
	}
	if ownerWrites != 1 || epochWrites != 1 || tenantWrites != 1 {
		t.Fatalf("legacy writes owner/epoch/tenant = %d/%d/%d",
			ownerWrites, epochWrites, tenantWrites)
	}
}
