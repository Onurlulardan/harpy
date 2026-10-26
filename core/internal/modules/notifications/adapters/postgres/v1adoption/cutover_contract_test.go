package v1adoption

import (
	"strings"
	"testing"
)

func TestCutoverContractPinsStaticCompositeApply(t *testing.T) {
	contract, err := NewCutoverContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ID() != CutoverContractID {
		t.Fatalf("contract ID = %q", contract.ID())
	}
	for _, fragment := range []string{
		"notifications_v1_adoption_apply_and_record_v1(",
		"$1::uuid", "$5::smallint", "$15::jsonb", "$19::text",
	} {
		if !strings.Contains(contract.Query(), fragment) {
			t.Fatalf("static contract query is missing %q", fragment)
		}
	}
	if strings.Contains(contract.Query(), "EXECUTE") || strings.Contains(contract.Query(), "%I") {
		t.Fatal("static contract query must not dispatch identifiers")
	}
}
