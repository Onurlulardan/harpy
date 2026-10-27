package formsv1

import "testing"

func TestFormsV1ConsumerContractIsVersionedAndStrict(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if contract.Validate() != nil || len(contract.Routes) != 2 ||
		contract.Routes[0].ID != "records.list" ||
		contract.Routes[1].ID != "records.export.legacy" {
		t.Fatal("Forms V1 consumer contract drifted")
	}
	contract.Routes[0].Errors[0].Code = "changed"
	fresh, err := Load()
	if err != nil || fresh.Routes[0].Errors[0].Code == "changed" {
		t.Fatal("Forms V1 consumer contract is not immutable per load")
	}
}
