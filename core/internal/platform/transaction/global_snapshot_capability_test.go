package transaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestGlobalSnapshotCapabilityRequiresExactAuthenticatedEvidence(t *testing.T) {
	issuer := NewGlobalSnapshotCapabilityIssuer()
	capability, err := issuer.Issue(
		uuid.New(), "billing.catalog.offering.list", "platform.billing.read",
		7, 5, 3, 2,
	)
	if err != nil || !capability.valid() {
		t.Fatalf("valid snapshot capability = %#v, %v", capability, err)
	}
	if _, err = issuer.Issue(
		uuid.New(), "billing.catalog.offering.list", "billing.read", 7, 5, 3, 2,
	); err == nil {
		t.Fatal("tenant permission minted global snapshot capability")
	}
}
