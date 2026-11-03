package baseline

import (
	"strings"
	"testing"
)

func TestBaselineScriptsAreGooseSnapshots(t *testing.T) {
	for name, script := range map[string]string{
		"platform":                      platformScript,
		"tenant":                        tenantScript,
		"platform cleanup":              platformCatalogRemovalScript,
		"platform tenant readiness":     platformTenantExecutorReadinessScript,
		"platform tenant readiness fix": platformTenantExecutorReadinessFixScript,
		"platform tenant claim fix":     platformTenantExecutorClaimFixScript,
		"tenant cleanup":                tenantCatalogRemovalScript,
	} {
		if strings.Contains(script, "\\restrict") || strings.Contains(script, "\\unrestrict") {
			t.Fatalf("%s baseline contains a psql meta command", name)
		}
		if !strings.Contains(script, "-- +goose Up") || !strings.Contains(script, "-- +goose StatementBegin") {
			t.Fatalf("%s baseline is not a goose statement", name)
		}
	}
	if !strings.Contains(tenantScript, tenantSchemaToken) ||
		!strings.Contains(tenantCatalogRemovalScript, tenantSchemaToken) {
		t.Fatal("tenant SQL scripts must contain the tenant schema token")
	}
	if strings.Contains(platformScript, tenantSchemaToken) ||
		strings.Contains(platformCatalogRemovalScript, tenantSchemaToken) {
		t.Fatal("platform SQL scripts must not contain the tenant schema token")
	}
}

func TestCombineBaselineResultsKeepsLatestPlatformVersion(t *testing.T) {
	fix := combineBaselineResults(
		Result{Applied: true, Version: platformTenantExecutorReadinessFixVersion},
		Result{Applied: true, Version: platformTenantExecutorClaimFixVersion},
	)
	applied := combineBaselineResults(
		Result{Applied: true, Version: platformVersion},
		Result{Applied: true, Version: platformCatalogRemovalVersion},
		Result{Applied: true, Version: platformTenantExecutorReadinessVersion},
		fix,
	)
	if !applied.Applied || applied.Version != platformTenantExecutorClaimFixVersion ||
		len(applied.AppliedVersions) != 5 ||
		applied.AppliedVersions[0] != platformVersion ||
		applied.AppliedVersions[1] != platformCatalogRemovalVersion ||
		applied.AppliedVersions[2] != platformTenantExecutorReadinessVersion ||
		applied.AppliedVersions[3] != platformTenantExecutorReadinessFixVersion ||
		applied.AppliedVersions[4] != platformTenantExecutorClaimFixVersion {
		t.Fatalf("applied baseline result = %+v", applied)
	}
	repeatedFix := combineBaselineResults(
		Result{Version: platformTenantExecutorReadinessFixVersion},
		Result{Version: platformTenantExecutorClaimFixVersion},
	)
	repeated := combineBaselineResults(
		Result{Version: platformVersion},
		Result{Version: platformCatalogRemovalVersion},
		Result{Version: platformTenantExecutorReadinessVersion},
		repeatedFix,
	)
	if repeated.Applied || repeated.Version != platformTenantExecutorClaimFixVersion ||
		len(repeated.AppliedVersions) != 0 {
		t.Fatalf("repeated baseline result = %+v", repeated)
	}
}

func TestPlatformMigrationSetIsOrderedAndIncludesReadinessCompatibility(t *testing.T) {
	migrations := PlatformMigrations()
	if len(migrations) != 5 ||
		migrations[0].Version != platformVersion ||
		migrations[1].Version != platformCatalogRemovalVersion ||
		migrations[2].Version != platformTenantExecutorReadinessVersion ||
		migrations[3].Version != platformTenantExecutorReadinessFixVersion ||
		migrations[4].Version != platformTenantExecutorClaimFixVersion ||
		LatestPlatformVersion() != platformTenantExecutorClaimFixVersion {
		t.Fatalf("platform baseline migration set = %+v", migrations)
	}
	for _, migration := range migrations {
		if migration.Path == "" || len(migration.SHA256) != 64 {
			t.Fatalf("invalid platform migration metadata = %+v", migration)
		}
	}
}

func TestTenantExecutorReadinessFixIgnoresGooseVersionZero(t *testing.T) {
	const totalApplied = "WHERE history.is_applied\n          ) = 3"
	if strings.Contains(platformTenantExecutorReadinessFixScript, totalApplied) {
		t.Fatal("readiness fix still requires every applied Goose row to total 3")
	}
	if !strings.Contains(platformTenantExecutorReadinessFixScript, "history.version_id <> 0") ||
		!strings.Contains(platformTenantExecutorReadinessFixScript, "20261028000000, 20261102000000, 20261103000000") {
		t.Fatal("readiness fix must require the three baseline versions and ignore Goose version 0")
	}
}

func TestTenantSchemaName(t *testing.T) {
	if validSchema("") || validSchema("Tenant") || validSchema("1abc") || validSchema(tenantSchemaToken) {
		t.Fatal("invalid schema names were accepted")
	}
	if !validSchema("tenant_a") {
		t.Fatal("tenant_a was rejected")
	}
}

func TestCatalogRemovalMigrationsFailClosedAndRetainIAMHistory(t *testing.T) {
	for name, script := range map[string]string{
		"platform": platformCatalogRemovalScript,
		"tenant":   tenantCatalogRemovalScript,
	} {
		for _, guard := range []string{
			"requires an explicit data disposition",
			"blocked by a foreign key",
			"blocked by a view",
			"blocked by a routine",
			"blocked by a trigger",
		} {
			if !strings.Contains(script, guard) {
				t.Errorf("%s cleanup is missing %q guard", name, guard)
			}
		}
	}
	for _, statement := range []string{
		"SET state = 'revoked'",
		"SET state = 'retired'",
		"iam_resources",
		"iam_permissions",
	} {
		if !strings.Contains(tenantCatalogRemovalScript, statement) {
			t.Errorf("tenant cleanup is missing %q", statement)
		}
	}
	if strings.Contains(tenantCatalogRemovalScript, "DELETE FROM __tenant_schema__.iam_resources") ||
		strings.Contains(tenantCatalogRemovalScript, "DELETE FROM __tenant_schema__.iam_permissions") {
		t.Fatal("tenant cleanup must retain IAM history instead of hard-deleting it")
	}
	if strings.Contains(platformCatalogRemovalScript, "CASCADE") ||
		strings.Contains(tenantCatalogRemovalScript, "CASCADE") {
		t.Fatal("Catalog cleanup must drop objects without CASCADE")
	}
}
