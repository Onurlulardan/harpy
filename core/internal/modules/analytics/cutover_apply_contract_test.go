package analytics

import "testing"

func TestCutoverApplyContractExposesAnalyticsDescriptor(t *testing.T) {
	contract, err := CutoverApplyContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ID() != "analytics.v1_adoption.v1" {
		t.Fatalf("Analytics cutover contract ID=%q", contract.ID())
	}
}
