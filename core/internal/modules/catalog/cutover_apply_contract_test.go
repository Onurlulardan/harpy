package catalog

import (
	"strings"
	"testing"
)

func TestCatalogCutoverFacadeUsesStaticCompositeContract(t *testing.T) {
	contract, err := CutoverApplyContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ID() != "catalog.v1_adoption.v1" ||
		!strings.Contains(contract.Query(), "catalog_v1_adoption_apply_and_record_v1") ||
		strings.Contains(contract.Query(), "modules") {
		t.Fatalf("unexpected Catalog apply contract %q", contract.ID())
	}
}
