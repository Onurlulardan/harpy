package v1authority

import (
	"strings"
	"testing"
)

func TestCutoverContractBindsProtectedAuthorityComposite(t *testing.T) {
	contract, err := NewCutoverContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ID() != "iam.v1_protected_authority_adoption.v1" {
		t.Fatalf("contract ID = %q", contract.ID())
	}
	if !strings.Contains(
		contract.Query(),
		"platform.iam_v1_protected_authority_apply_and_record_v1",
	) {
		t.Fatal("IAM V1 protected authority composite query is not bound")
	}
}
