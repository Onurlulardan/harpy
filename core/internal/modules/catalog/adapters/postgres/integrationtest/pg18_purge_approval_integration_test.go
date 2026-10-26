//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestCatalogPG18PurgeApprovalAtomicityReplayAndIsolation(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startCatalogPG18(t, ctx, "approval_"+origin)
			prepareCatalogPG18(t, ctx, database, origin)
			coordinator := openCatalogCoordinator(t, ctx, database)
			assertWorkerTwoOwnerRollbackAndSettlement(t, ctx, database, coordinator)
			approvedAt := time.Date(2026, 7, 31, 8, 0, 0, 123000, time.UTC)
			receipt := catalogApprovalReceipt(t, approvedAt)
			seedApprovalInstallation(t, ctx, database, approvedAt)
			assertApprovalCreateAndReplay(t, ctx, database, coordinator, receipt)
			assertForgedPurgeRollsBack(t, ctx, database, coordinator, receipt)
			assertPurgeConsumesAndReplays(t, ctx, database, coordinator, receipt)
			assertApprovalCannotCrossTenant(t, ctx, coordinator, receipt)
		})
	}
}

func assertApprovalCreateAndReplay(t *testing.T, ctx context.Context,
	database catalogPG18, coordinator *platformtx.Coordinator,
	receipt catalogapp.PurgeApprovalReceipt,
) {
	t.Helper()
	request := catalogLifecycleRequest(t, catalogTenantA, "catalog.installation.purge.approve")
	stored, replayed := approveCatalogPurge(t, ctx, coordinator, request, receipt)
	if replayed || stored != receipt {
		t.Fatalf("approval create replay=%v stored=%+v", replayed, stored)
	}
	retry := receipt
	retry.ID = "019fa100-0000-7000-8000-000000000008"
	retry.ApprovalSHA256 = approvalDigest(
		"catalog.purge-approval-receipt.v2", retry.ID, retry.TenantID,
		retry.InstallationID, retry.PackageID, retry.OperationID,
		retry.RequestedBy, retry.ApprovedBy, retry.BackupReceiptID,
		retry.BackupSHA256, retry.ConfirmationSlug, retry.Reason,
		retry.PurgeIdempotencyKey, retry.RequestSHA256, retry.ApprovalIdempotencyKey,
		approvalTime(retry.StepUpAuthenticatedAt), approvalTime(retry.ApprovedAt),
		approvalTime(retry.ExpiresAt),
	)
	stored, replayed = approveCatalogPurge(t, ctx, coordinator, request, retry)
	if !replayed || stored.ID != receipt.ID {
		t.Fatalf("approval replay=%v stored=%+v", replayed, stored)
	}
	assertApprovalReceiptImmutable(t, ctx, database)
	forged := retry
	forged.BackupReceiptID = "019fa100-0000-7000-8000-000000000099"
	if _, _, err := tryApproveCatalogPurge(ctx, coordinator, request, forged); err == nil {
		t.Fatal("approval idempotency drift was accepted")
	}
}

func approveCatalogPurge(t *testing.T, ctx context.Context,
	coordinator *platformtx.Coordinator, request platformtx.ScopeRequest,
	receipt catalogapp.PurgeApprovalReceipt,
) (catalogapp.PurgeApprovalReceipt, bool) {
	t.Helper()
	value, replayed, err := tryApproveCatalogPurge(ctx, coordinator, request, receipt)
	if err != nil {
		t.Fatal(err)
	}
	return value, replayed
}

func tryApproveCatalogPurge(ctx context.Context, coordinator *platformtx.Coordinator,
	request platformtx.ScopeRequest, receipt catalogapp.PurgeApprovalReceipt,
) (stored catalogapp.PurgeApprovalReceipt, replayed bool, err error) {
	err = coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		store, _, bindErr := lifecyclepostgres.NewBinder().Bind(
			scoped, scope, catalogapp.ScopeCommand,
		)
		if bindErr != nil {
			return bindErr
		}
		writer := store.(catalogapp.PurgeApprovalStore)
		stored, replayed, bindErr = writer.ApprovePurge(
			scoped, catalogapp.PurgeApprovalMutation{Receipt: receipt},
		)
		return bindErr
	})
	return
}

func seedApprovalInstallation(t *testing.T, ctx context.Context,
	database catalogPG18, at time.Time,
) {
	t.Helper()
	_, err := database.owner.ExecContext(ctx, `INSERT INTO catalog_a.catalog_installations(
installation_id,package_id,active_version,state,reference_epoch,row_version,updated_at)
VALUES($1,$2,'1.0.0','RETIRED',2,3,$3)`,
		approvalInstallationID, approvalPackageID, at)
	if err != nil {
		t.Fatal(err)
	}
}

func assertForgedPurgeRollsBack(t *testing.T, ctx context.Context,
	database catalogPG18, coordinator *platformtx.Coordinator,
	receipt catalogapp.PurgeApprovalReceipt,
) {
	t.Helper()
	assertStalePurgeRejected(t, ctx, database, coordinator, receipt)
	at := receipt.ApprovedAt.Add(5 * time.Minute)
	forged := catalogPurgeMutation(t, receipt,
		"019fa100-0000-7000-8000-000000000099", approvalOperationID, at, true)
	request := catalogLifecycleRequestForActor(
		t, catalogTenantA, approvalRequesterID, "catalog.installation.purge",
	)
	if _, err := createCatalogPurge(ctx, coordinator, request, forged); err == nil {
		t.Fatal("forged backup approval was accepted")
	}
	assertCatalogRowCount(t, ctx, database.owner, "catalog_a.catalog_operation_runs", 0)
	assertApprovalState(t, ctx, database, "RETIRED", "AVAILABLE")
}

func assertPurgeConsumesAndReplays(t *testing.T, ctx context.Context,
	database catalogPG18, coordinator *platformtx.Coordinator,
	receipt catalogapp.PurgeApprovalReceipt,
) {
	t.Helper()
	at := receipt.ApprovedAt.Add(5 * time.Minute)
	mutation := catalogPurgeMutation(t, receipt, approvalBackupID, approvalOperationID, at, true)
	request := catalogLifecycleRequestForActor(
		t, catalogTenantA, approvalRequesterID, "catalog.installation.purge",
	)
	handle, err := createCatalogPurge(ctx, coordinator, request, mutation)
	if err != nil || handle.Replayed {
		t.Fatalf("purge create replay=%v err=%v", handle.Replayed, err)
	}
	assertApprovalState(t, ctx, database, "PURGE_PENDING", "CONSUMED")
	handle, err = createCatalogPurge(ctx, coordinator, request, mutation)
	if err != nil || !handle.Replayed {
		t.Fatalf("purge exact replay=%v err=%v", handle.Replayed, err)
	}
	reused := catalogPurgeMutation(t, receipt, approvalBackupID,
		"019fa100-0000-7000-8000-000000000009", at.Add(time.Microsecond), false)
	reused.IdempotencyKey = "purge-proof-0002"
	reused.PurgeAuthority.IdempotencyKey = reused.IdempotencyKey
	reused.PurgeAuthority.AuthoritySHA256 =
		catalogappAuthoritySHA(*reused.PurgeAuthority)
	reused.RequestSHA256 = reused.PurgeAuthority.AuthoritySHA256
	if _, err = createCatalogPurge(ctx, coordinator, request, reused); err == nil {
		t.Fatal("consumed approval was reused by another operation")
	}
	assertCatalogRowCount(t, ctx, database.owner, "catalog_a.catalog_operation_runs", 1)
}

func createCatalogPurge(ctx context.Context, coordinator *platformtx.Coordinator,
	request platformtx.ScopeRequest, mutation catalogapp.StartMutation,
) (handle catalogapp.OperationHandle, err error) {
	err = coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		handle, err = createCatalogOperation(scoped, scope, mutation)
		return err
	})
	return
}

func assertApprovalCannotCrossTenant(t *testing.T, ctx context.Context,
	coordinator *platformtx.Coordinator, receipt catalogapp.PurgeApprovalReceipt,
) {
	t.Helper()
	request := catalogLifecycleRequest(t, catalogTenantB, "catalog.installation.purge.approve")
	err := coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		store, _, bindErr := lifecyclepostgres.NewBinder().Bind(
			scoped, scope, catalogapp.ScopeSnapshot,
		)
		if bindErr != nil {
			return bindErr
		}
		_, bindErr = store.(catalogapp.PurgeApprovalStore).ApprovalReceipt(
			scoped, uuid.MustParse(receipt.ID),
		)
		return bindErr
	})
	if err == nil {
		t.Fatal("cross-tenant approval receipt read was accepted")
	}
}
