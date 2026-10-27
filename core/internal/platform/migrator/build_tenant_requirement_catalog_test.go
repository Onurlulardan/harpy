package migrator

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func TestBuildTenantRequirementCatalogBindsSignedSelectionAndRelease(t *testing.T) {
	runner := tenantRequirementCatalogFixture()
	targets := tenantRequirementTargetFixture()
	catalog, err := runner.buildTenantRequirementCatalog(runnerTestRunID, targets)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.MigrationRunID != runnerTestRunID || catalog.ReleaseID != runner.identity.releaseID ||
		catalog.MigrationHash != runner.identity.manifestSHA256 ||
		catalog.RequirementCount != 2 || len(catalog.Requirements) != 2 ||
		catalog.TargetCount != 2 || len(catalog.Targets) != 2 {
		t.Fatalf("tenant requirement catalog = %#v", catalog)
	}
	for index, requirement := range catalog.Requirements {
		if requirement.Ordinal != index+1 ||
			requirement.Version != runner.tenantSelected[index].Version ||
			requirement.SelectedHash != runner.tenantSelected[index].SHA256 {
			t.Fatalf("requirement %d = %#v", index, requirement)
		}
	}
	digest := sha256.Sum256(catalog.Canonical)
	if catalog.CatalogHash != hex.EncodeToString(digest[:]) {
		t.Fatalf("catalog digest = %s", catalog.CatalogHash)
	}
	changed := runner
	changed.identity.releaseID = "018f0000-0000-7000-8000-000000000099"
	second, err := changed.buildTenantRequirementCatalog(runnerTestRunID, targets)
	if err != nil || second.CatalogHash == catalog.CatalogHash {
		t.Fatalf("release-bound catalog digest = %s err=%v", second.CatalogHash, err)
	}
}

func tenantRequirementTargetFixture() []postgresstore.TenantMigrationTarget {
	return []postgresstore.TenantMigrationTarget{
		{TenantID: "018f0000-0000-7000-8000-000000000001",
			SchemaName: "tenant_one"},
		{TenantID: "018f0000-0000-7000-8000-000000000002",
			SchemaName: "tenant_two"},
	}
}

func tenantRequirementCatalogFixture() *Runner {
	digest := strings.Repeat("a", 64)
	return &Runner{
		identity: runIdentity{
			releaseID:      "018f0000-0000-7000-8000-000000000091",
			environmentID:  "018f0000-0000-7000-8000-000000000092",
			databaseOrigin: postgresv1.DatabaseOriginFresh,
			manifestSHA256: digest,
		},
		tenantSelected: []postgresv1.SelectedMigration{
			{Owner: "tenancy", Scope: "tenant", Version: 20260714011000,
				Origin: postgresv1.DatabaseOriginFresh, Path: "tenant/first.sql", SHA256: digest},
			{Owner: "tenancy", Scope: "tenant", Version: 20260714011100,
				Origin: postgresv1.DatabaseOriginFresh, Path: "tenant/second.sql",
				SHA256: strings.Repeat("b", 64)},
		},
	}
}
