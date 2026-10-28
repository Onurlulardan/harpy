package storagecatalog

import (
	"path/filepath"
	"testing"
)

func TestRepositoryPlatformAsyncModularOwnerIsComplete(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	owner := repositoryOwner(t, manifest, "platform_async")
	requireCatalogCount(t, "schema grants", len(owner.SchemaUsageGrants), 3)
	assertCatalogBoundary(t, "schema grants",
		owner.SchemaUsageGrants[0].Role, "harpy_adopter",
		owner.SchemaUsageGrants[2].Role, "harpy_runtime")
	requireCatalogCount(t, "routines", len(owner.PlatformRoutines), 149)
	assertCatalogBoundary(t, "routines",
		owner.PlatformRoutines[0], "platform.platform_async_accept_v1_adoption_v1",
		owner.PlatformRoutines[148],
		"platform.platform_async_verify_stale_job_disposition_replay_v1")
	requireCatalogCount(t, "routine grants", len(owner.RoutineExecGrants), 72)
	assertCatalogBoundary(t, "routine grants",
		owner.RoutineExecGrants[0].Routine,
		"platform.platform_async_accept_v1_adoption_v1",
		owner.RoutineExecGrants[71].Routine,
		"platform.platform_async_validate_job_transition_v2")
	requireCatalogCount(t, "tables", len(owner.PlatformTables), 15)
	assertCatalogBoundary(t, "tables",
		owner.PlatformTables[0], "platform.artifact_quota_counters",
		owner.PlatformTables[14],
		"platform.platform_async_installation_activation_receipts")
	requireCatalogCount(t, "foreign keys", len(owner.ForeignKeys), 29)
	assertCatalogBoundary(t, "foreign keys",
		owner.ForeignKeys[0].Constraint, "artifact_quota_counters_tenant_fk",
		owner.ForeignKeys[28].Constraint, "outbox_tenant_fk")
	requireCatalogCount(t, "external accesses", len(owner.ExternalAccesses), 12)
	assertCatalogBoundary(t, "external accesses",
		owner.ExternalAccesses[0].Target,
		"platform.billing_resolve_current_tenant_authority_receipt_v1",
		owner.ExternalAccesses[11].Target, "platform.tenants")
}

func repositoryOwner(t *testing.T, manifest Manifest, name string) Owner {
	t.Helper()
	for _, owner := range manifest.Owners {
		if owner.Name == name {
			return owner
		}
	}
	t.Fatalf("repository owner %q is missing", name)
	return Owner{}
}

func assertCatalogBoundary(
	t *testing.T, label, first, wantFirst, last, wantLast string,
) {
	t.Helper()
	if first != wantFirst || last != wantLast {
		t.Fatalf("%s boundary = %q/%q", label, first, last)
	}
}

func requireCatalogCount(t *testing.T, label string, count, want int) {
	t.Helper()
	if count != want {
		t.Fatalf("%s count = %d, want %d", label, count, want)
	}
}
