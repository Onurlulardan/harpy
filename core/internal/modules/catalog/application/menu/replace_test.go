package menu

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReplaceRejectsMissingRevision(t *testing.T) {
	_, err := (*Application)(nil).Replace(
		context.Background(), platformtx.TenantScope{},
		ReplaceRequest{IdempotencyKey: "catalog-menu-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("replace menu error = %v", err)
	}
}

func TestReplaceDigestEmptyMatchesBaseline(t *testing.T) {
	if replaceDigest(nil) != emptyMenuDigest {
		t.Fatal("empty tenant menu must keep the catalog empty digest")
	}
}

func TestActivateTenantAuthoredMarksReservedPackage(t *testing.T) {
	active := activateTenantAuthored(nil)
	if !active[TenantAuthoredPackageID] {
		t.Fatal("tenant-authored menu items must remain visible without a package install")
	}
}
