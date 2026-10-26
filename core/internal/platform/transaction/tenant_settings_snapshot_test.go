package transaction

import "testing"

func TestTenantSettingsSnapshotIssuerOwnsExactFiveDomains(t *testing.T) {
	operation, err := NewTenantSettingsSnapshotIssuer().Issue(
		"tenancy.settings.get",
		map[string]int64{"notifications": 4, "tenancy": 1, "iam": 2,
			"forms": 5, "billing": 3},
	)
	if err != nil || !operation.valid() || operation.mode != SnapshotQuery ||
		len(operation.domains) != 5 || !operation.allowsOwner("tenancy") ||
		!operation.allowsOwner("billing") || operation.allowsOwner("work") {
		t.Fatalf("tenant settings snapshot = %+v err=%v", operation, err)
	}
}

func TestTenantSettingsSnapshotIssuerRejectsRESTQueryOwnerSets(t *testing.T) {
	issuer := NewTenantSettingsSnapshotIssuer()
	if _, err := issuer.Issue("restquery.execute", map[string]int64{
		"billing": 1, "forms": 1, "iam": 1, "notifications": 1, "tenancy": 1,
	}); err == nil {
		t.Fatal("REST Query operation was accepted as tenant settings snapshot")
	}
	if _, err := issuer.Issue("tenancy.settings.get", map[string]int64{
		"billing": 1, "forms": 1, "iam": 1, "notifications": 1,
	}); err == nil {
		t.Fatal("four-owner snapshot was accepted")
	}
	if _, err := NewSnapshotAuthorityIssuer().Issue("restquery.execute", map[string]int64{
		"billing": 1, "forms": 1, "iam": 1, "notifications": 1, "tenancy": 1,
	}); err == nil {
		t.Fatal("REST Query issuer accepted five owners")
	}
}
