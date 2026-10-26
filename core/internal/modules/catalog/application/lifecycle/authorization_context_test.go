package lifecycle

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func TestAuthorizationReceiptIsBoundToExactPurgeOperation(t *testing.T) {
	receipt := AuthorizationReceipt{
		actorID:    uuid.MustParse("019fa000-0000-7000-8000-000000000001"),
		tenantID:   uuid.MustParse("019fa000-0000-7000-8000-000000000002"),
		permission: "catalog.purge", operation: operationPurge,
		tokenVersion: 1, authorizationVersion: 2, authorizationClock: 3,
		permissionID:        uuid.MustParse("019fa000-0000-7000-8000-000000000003"),
		evidenceSHA256:      [sha256.Size]byte{1},
		stepUpPurposeSHA256: [sha256.Size]byte{2},
	}
	if !receipt.valid() {
		t.Fatal("exact Catalog purge authorization receipt is invalid")
	}
	receipt.operation = operationPurgeApprove
	if receipt.valid() {
		t.Fatal("purge approval operation reused a purge execution receipt")
	}
}
