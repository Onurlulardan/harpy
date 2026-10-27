package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestApprovePurgeRejectsEmptyAuthority(t *testing.T) {
	_, _, err := (&Application{}).ApprovePurge(
		context.Background(), platformtx.TenantScope{}, PurgeApprovalRequest{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty approval error = %v", err)
	}
}

func TestPurgeApprovalAcceptsOpaqueStablePackageIdentity(t *testing.T) {
	value, err := uuidFromCanonical("e3000000-0000-0000-0000-000000000001")
	if err != nil || value.String() != "e3000000-0000-0000-0000-000000000001" {
		t.Fatalf("opaque stable package identity = %s, %v", value, err)
	}
}
