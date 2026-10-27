//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func assertStalePurgeRejected(
	t *testing.T, ctx context.Context, database catalogPG18,
	coordinator *platformtx.Coordinator, receipt catalogapp.PurgeApprovalReceipt,
) {
	t.Helper()
	stale := catalogPurgeMutation(
		t, receipt, approvalBackupID, approvalOperationID,
		receipt.ExpiresAt.Add(time.Microsecond), true,
	)
	request := catalogLifecycleRequestForActor(
		t, catalogTenantA, approvalRequesterID, "catalog.installation.purge",
	)
	if _, err := createCatalogPurge(ctx, coordinator, request, stale); err == nil {
		t.Fatal("expired purge approval was accepted")
	}
	assertCatalogRowCount(t, ctx, database.owner, "catalog_a.catalog_operation_runs", 0)
	assertApprovalState(t, ctx, database, "RETIRED", "AVAILABLE")
}
