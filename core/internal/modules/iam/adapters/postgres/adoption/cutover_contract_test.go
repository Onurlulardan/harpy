package adoption

import (
	"strings"
	"testing"
)

func TestCutoverApplyContractBindsExactIAMRoutine(t *testing.T) {
	contract, err := NewCutoverContract()
	if err != nil {
		t.Fatalf("create IAM cutover apply contract: %v", err)
	}
	if contract.ID() != ContractID ||
		!strings.Contains(contract.Query(), "iam_v1_adoption_apply_and_record_v1") ||
		strings.Contains(contract.Query(), "users_form_table") {
		t.Fatalf("unexpected IAM cutover apply contract: %s", contract.ID())
	}
}
