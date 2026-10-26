package demomaterialization

import (
	"errors"
	"testing"
)

func TestVerifyRequestRequiresPackageBoundSourceContract(t *testing.T) {
	request := fixtureRequest(t)
	if _, err := verifyRequest(request); err != nil {
		t.Fatalf("verify valid Catalog demo request: %v", err)
	}
	request.PackageSlug = "sales-crm"
	if _, err := verifyRequest(request); !errors.Is(err, ErrInvalid) {
		t.Fatalf("package/source mismatch error=%v", err)
	}
}

func TestVerifyRequestRejectsCatalogContractDrift(t *testing.T) {
	request := fixtureRequest(t)
	request.Component.Contract = "catalog.demo-data.v2"
	if _, err := verifyRequest(request); !errors.Is(err, ErrInvalid) {
		t.Fatalf("contract drift error=%v", err)
	}
}
