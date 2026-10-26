package v1adoption

import (
	"strings"
	"testing"
)

func TestCutoverApplyContractBindsExactWorkRoutine(t *testing.T) {
	contract, err := NewCutoverContract()
	if err != nil {
		t.Fatalf("create Work cutover apply contract: %v", err)
	}
	if contract.ID() != CutoverContractID ||
		!strings.Contains(contract.Query(), "work_v1_adoption_apply_and_record_v1") ||
		strings.Contains(contract.Query(), "task_boards") {
		t.Fatalf("unexpected Work cutover apply contract: %s", contract.ID())
	}
}
