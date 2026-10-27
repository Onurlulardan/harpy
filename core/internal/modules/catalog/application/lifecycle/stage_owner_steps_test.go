package lifecycle

import (
	"testing"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func TestOwnerStepCommandCarriesEpochOnlyForPurge(t *testing.T) {
	operationID, err := packagecontract.ParseStableID(
		"019fa000-0000-7000-8000-000000000301",
	)
	if err != nil {
		t.Fatal(err)
	}
	componentID, err := packagecontract.ParseComponentID(
		"019fa000-0000-7000-8000-000000000302",
	)
	if err != nil {
		t.Fatal(err)
	}
	operation := catalogdomain.Operation{ID: operationID}
	pin := catalogdomain.ComponentPin{
		ComponentID: componentID, Owner: "analytics", OwnerGeneration: 1,
		Contract:      "analytics.catalog-definition.v1",
		ContentSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	purge := ownerStepCommand(operation, pin, "purge", 7)
	if purge.ReferenceEpoch != 7 {
		t.Fatalf("purge reference epoch = %d, want 7", purge.ReferenceEpoch)
	}
	stage := ownerStepCommand(operation, pin, "stage", 7)
	if stage.ReferenceEpoch != 0 {
		t.Fatalf("stage reference epoch = %d, want 0", stage.ReferenceEpoch)
	}
}
