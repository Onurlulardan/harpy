package catalog

import (
	"testing"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
)

func TestLifecycleAuthorizationReceiptIsAttachedOnlyToExactPurge(t *testing.T) {
	if !attachLifecycleAuthorizationReceipt(
		cataloghttp.PermissionPurge, "catalog.installation.purge",
	) {
		t.Fatal("exact Catalog purge lost its lifecycle authorization receipt")
	}
	if attachLifecycleAuthorizationReceipt(
		cataloghttp.PermissionPurge, "catalog.installation.purge.approve",
	) {
		t.Fatal("Catalog purge approval received a purge execution receipt")
	}
}
