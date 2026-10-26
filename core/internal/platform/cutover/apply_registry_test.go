package cutover

import "testing"

func TestApplyRegistryRequiresUniqueTypedContracts(t *testing.T) {
	contract, err := NewApplyContract("iam.v1_adoption.v1", "SELECT $1")
	if err != nil {
		t.Fatalf("create apply contract: %v", err)
	}
	registry, err := NewApplyRegistry(contract)
	if err != nil {
		t.Fatalf("create apply registry: %v", err)
	}
	query, found := registry.Lookup(contract.ID())
	if !found || query != "SELECT $1" || !registry.Valid() {
		t.Fatalf("apply contract lookup = %q, %t", query, found)
	}
	if _, err := NewApplyRegistry(contract, contract); err == nil {
		t.Fatal("duplicate apply contract was accepted")
	}
}

func TestApplyContractRejectsUntypedIdentity(t *testing.T) {
	for _, id := range []string{"", "IAM.v1", "iam v1", ".iam"} {
		if _, err := NewApplyContract(id, "SELECT 1"); err == nil {
			t.Fatalf("invalid contract ID %q was accepted", id)
		}
	}
	if _, err := NewApplyContract("iam.v1", " "); err == nil {
		t.Fatal("empty apply query was accepted")
	}
}
