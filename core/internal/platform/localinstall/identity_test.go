package localinstall

import (
	"testing"

	"github.com/google/uuid"
)

func TestIdentityIsStableAndPurposeSeparated(t *testing.T) {
	environment := uuid.Must(uuid.NewV7())
	first, second := DeriveIdentity(environment), DeriveIdentity(environment)
	if first != second || first.LicenseID.Version() != uuid.Version(7) ||
		first.LicenseID == first.OfferingVersionID {
		t.Fatalf("local identity drifted: first=%+v second=%+v", first, second)
	}
	if NextCredentialID(first.RefreshOperationID, 2) ==
		NextCredentialID(first.RefreshOperationID, 3) {
		t.Fatal("credential versions share an identity")
	}
}
