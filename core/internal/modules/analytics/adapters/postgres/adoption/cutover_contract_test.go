package adoption

import (
	"strings"
	"testing"
)

func TestCutoverApplyContractBindsExactAnalyticsRoutine(t *testing.T) {
	contract, err := NewCutoverContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ID() != CutoverContractID ||
		!strings.Contains(contract.Query(), "analytics_v1_adoption_apply_and_record_v1") ||
		strings.Count(contract.Query(), "$") != 19 {
		t.Fatalf("unexpected Analytics cutover contract: %q %q",
			contract.ID(), contract.Query())
	}
}
