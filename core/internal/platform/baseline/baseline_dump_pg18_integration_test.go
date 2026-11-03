package baseline

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestBaselineDumpCleanupPostgres18(t *testing.T) {
	dsn := os.Getenv(baselinePG18TestDSNKey)
	if dsn == "" {
		t.Skipf("%s is not set", baselinePG18TestDSNKey)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	const tenantSchema = "tenant_baseline_dump"
	resetDumpSchemas(t, ctx, dsn, tenantSchema)
	createDumpRoles(t, ctx, dsn)

	platformResult, err := ApplyPlatform(ctx, dsn)
	if err != nil {
		t.Fatalf("apply platform baseline: %v", err)
	}
	if !platformResult.Applied || platformResult.Version != platformTenantExecutorCompletionVersion ||
		len(platformResult.AppliedVersions) != 4 ||
		platformResult.AppliedVersions[0] != platformVersion ||
		platformResult.AppliedVersions[1] != platformCatalogRemovalVersion ||
		platformResult.AppliedVersions[2] != platformTenantExecutorBaselineVersion ||
		platformResult.AppliedVersions[3] != platformTenantExecutorCompletionVersion {
		t.Fatalf("platform baseline result = %+v", platformResult)
	}
	assertBaselineTenantExecutorCanRegister(t, ctx, dsn)
	tenantResult, err := ApplyTenant(ctx, dsn, tenantSchema)
	if err != nil {
		t.Fatalf("apply tenant baseline: %v", err)
	}
	if !tenantResult.Applied || tenantResult.Version != 20261102000001 ||
		len(tenantResult.AppliedVersions) != 2 ||
		tenantResult.AppliedVersions[0] != tenantVersion ||
		tenantResult.AppliedVersions[1] != 20261102000001 {
		t.Fatalf("tenant baseline result = %+v", tenantResult)
	}
	assertDumpCleanup(t, ctx, dsn, tenantSchema)

	repeatedPlatform, err := ApplyPlatform(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	repeatedTenant, err := ApplyTenant(ctx, dsn, tenantSchema)
	if err != nil {
		t.Fatal(err)
	}
	if repeatedPlatform.Applied || repeatedTenant.Applied ||
		repeatedPlatform.Version != platformTenantExecutorCompletionVersion ||
		repeatedTenant.Version != 20261102000001 ||
		len(repeatedPlatform.AppliedVersions) != 0 ||
		len(repeatedTenant.AppliedVersions) != 0 {
		t.Fatalf("repeated apply = platform %+v tenant %+v", repeatedPlatform, repeatedTenant)
	}
}

func assertBaselineTenantExecutorCanRegister(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(ctx)
	transaction, err := connection.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	if _, err := transaction.Exec(ctx, "SET LOCAL ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}

	const profileID = "11111111-1111-4111-8111-111111111111"
	const releaseID = "22222222-2222-4222-8222-222222222222"
	const manifestSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var actualProfileID, actualReleaseID, environmentID, deploymentMode string
	var runtimeGeneration, databaseGeneration int64
	var actualManifestSHA256 string
	var readyUntil time.Time
	err = transaction.QueryRow(ctx,
		"SELECT profile_instance_id, release_id, environment_id, deployment_mode, "+
			"runtime_generation, database_generation, migration_manifest_sha256, ready_until "+
			"FROM platform.tenancy_tenant_migration_ready_register_v1($1::uuid, $2::uuid, 'onpremise', 1, $3, 30)",
		profileID, releaseID, manifestSHA256,
	).Scan(&actualProfileID, &actualReleaseID, &environmentID, &deploymentMode,
		&runtimeGeneration, &databaseGeneration, &actualManifestSHA256, &readyUntil)
	if err != nil {
		t.Fatal(err)
	}
	if actualProfileID != profileID || actualReleaseID != releaseID ||
		environmentID == "" || deploymentMode != "onpremise" ||
		runtimeGeneration != 1 || databaseGeneration != 1 ||
		actualManifestSHA256 != manifestSHA256 || !readyUntil.After(time.Now()) {
		t.Fatalf("baseline readiness receipt differs: profile=%s release=%s env=%s mode=%s runtime=%d database=%d hash=%s ready_until=%s",
			actualProfileID, actualReleaseID, environmentID, deploymentMode,
			runtimeGeneration, databaseGeneration, actualManifestSHA256, readyUntil)
	}
}
