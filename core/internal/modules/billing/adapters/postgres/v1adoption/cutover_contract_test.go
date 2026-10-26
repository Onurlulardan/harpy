package v1adoption

import (
	"strings"
	"testing"
)

func TestCutoverContractsBindExactBillingRoutines(t *testing.T) {
	tests := []struct {
		name, id, routine, forbidden string
		build                        func() (CutoverContract, error)
	}{
		{"global", GlobalCutoverContractID,
			"billing_v1_adoption_global_apply_and_record_v1",
			"billing_v1_adoption_tenant_apply_and_record_v1", NewGlobalCutoverContract},
		{"tenant", TenantCutoverContractID,
			"billing_v1_adoption_tenant_apply_and_record_v1",
			"billing_v1_adoption_global_apply_and_record_v1", NewTenantCutoverContract},
		{"base", BaseCutoverContractID,
			"billing_v1_base_adoption_apply_and_record_v1",
			"billing_v1_adoption_global_apply_and_record_v1", NewBaseCutoverContract},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract, err := test.build()
			if err != nil {
				t.Fatal(err)
			}
			if contract.ID() != test.id ||
				!strings.Contains(contract.Query(), test.routine) ||
				strings.Contains(contract.Query(), test.forbidden) ||
				strings.Contains(contract.Query(), "billing_v1_adoption_runs") {
				t.Fatalf("unexpected Billing %s cutover contract", test.name)
			}
		})
	}
}
