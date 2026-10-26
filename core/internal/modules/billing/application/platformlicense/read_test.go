package platformlicense

import (
	"errors"
	"testing"
)

func TestReadReturnsValidatedLicense(t *testing.T) {
	environmentID := testV7(t)
	store := &storeFixture{license: validLicenseFixture(t, environmentID)}
	service, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	value, err := service.Read(t.Context(), environmentID)
	if err != nil || value.EnvironmentID != environmentID ||
		store.environmentID != environmentID {
		t.Fatalf("license=%+v environment=%s err=%v", value, store.environmentID, err)
	}
}

func TestReadRejectsMalformedStoreResult(t *testing.T) {
	environmentID := testV7(t)
	store := &storeFixture{license: validLicenseFixture(t, environmentID)}
	store.license.Features = nil
	service, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Read(t.Context(), environmentID); !errors.Is(
		err, ErrUnavailable,
	) {
		t.Fatalf("error=%v", err)
	}
}
