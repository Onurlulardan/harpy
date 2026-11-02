package baseline

import (
	"context"
	"os"
	"testing"
	"time"
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
	if !platformResult.Applied || platformResult.Version != 20261102000000 ||
		len(platformResult.AppliedVersions) != 2 ||
		platformResult.AppliedVersions[0] != platformVersion ||
		platformResult.AppliedVersions[1] != 20261102000000 {
		t.Fatalf("platform baseline result = %+v", platformResult)
	}
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
		repeatedPlatform.Version != 20261102000000 ||
		repeatedTenant.Version != 20261102000001 ||
		len(repeatedPlatform.AppliedVersions) != 0 ||
		len(repeatedTenant.AppliedVersions) != 0 {
		t.Fatalf("repeated apply = platform %+v tenant %+v", repeatedPlatform, repeatedTenant)
	}
}
