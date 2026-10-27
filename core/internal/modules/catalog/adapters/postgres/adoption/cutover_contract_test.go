package adoption

import (
	"strings"
	"testing"
)

func TestCutoverApplyContractBindsExactCatalogRoutine(t *testing.T) {
	contract, err := NewCutoverContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ID() != CutoverContractID ||
		!strings.Contains(contract.Query(), "catalog_v1_adoption_apply_and_record_v1") ||
		strings.Contains(contract.Query(), "modules") {
		t.Fatalf("unexpected Catalog cutover contract: %s", contract.ID())
	}
}
