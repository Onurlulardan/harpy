package transaction

import "testing"

func TestTenancyAuthorityGroupRequiresExactSingleton(t *testing.T) {
	issuer := NewAuthorityIssuer()
	operation, err := issuer.Issue(
		"tenancy.settings.update", "tenancy", 7, map[string]int64{"tenancy": 11},
	)
	if err != nil || !operation.valid() || !operation.allowsOwner("tenancy") ||
		operation.allowsOwner("iam") {
		t.Fatalf("tenancy authority rejected or over-broad: operation=%+v err=%v",
			operation, err)
	}
	if _, err = issuer.Issue(
		"tenancy.settings.update", "tenancy", 7,
		map[string]int64{"tenancy": 11, "iam": 13},
	); err == nil {
		t.Fatal("tenancy authority accepted an extra domain")
	}
}
