package v1environment

import (
	"strings"
	"testing"
)

func TestCutoverContractBindsTenancyV1EnvironmentComposite(t *testing.T) {
	contract, err := NewCutoverContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ID() != "tenancy.v1_environment_adoption.v1" {
		t.Fatalf("contract ID = %q", contract.ID())
	}
	if !strings.Contains(
		contract.Query(),
		"platform.tenancy_v1_environment_apply_and_record_v1",
	) {
		t.Fatal("tenancy V1 environment composite query is not bound")
	}
}
