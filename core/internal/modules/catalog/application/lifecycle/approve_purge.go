package lifecycle

import (
	"context"
	"time"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) ApprovePurge(
	ctx context.Context,
	scope platformtx.TenantScope,
	request PurgeApprovalRequest,
) (PurgeApprovalReceipt, bool, error) {
	if !validApprovalRequest(request) {
		return PurgeApprovalReceipt{}, false, ErrInvalid
	}
	store, identity, err := application.bind(
		ctx, scope, ScopeCommand, operationPurgeApprove, ActionPurge,
	)
	if err != nil {
		return PurgeApprovalReceipt{}, false, err
	}
	if identity.ActorID == request.RequestedBy {
		return PurgeApprovalReceipt{}, false, ErrForbidden
	}
	reader, ok := store.(InstallationByIDStore)
	if !ok || reader == nil {
		return PurgeApprovalReceipt{}, false, ErrUnavailable
	}
	installation, err := reader.InstallationByID(ctx, request.InstallationID)
	if err != nil || installation.ID.String() != request.InstallationID.String() {
		return PurgeApprovalReceipt{}, false, ErrNotFound
	}
	if installation.State != catalogdomain.InstallationRetired {
		return PurgeApprovalReceipt{}, false, ErrConflict
	}
	now, err := application.now()
	stepUpAt := time.Unix(request.StepUpAtUnix, 0).UTC()
	if err != nil || stepUpAt.After(now) || now.Sub(stepUpAt) > 10*time.Minute {
		return PurgeApprovalReceipt{}, false, ErrForbidden
	}
	packageID, err := uuidFromCanonical(installation.PackageID.String())
	if err != nil {
		return PurgeApprovalReceipt{}, false, ErrUnavailable
	}
	registry, err := application.registry.Bind(ctx, scope)
	if err != nil || registry == nil {
		return PurgeApprovalReceipt{}, false, ErrUnavailable
	}
	registered, err := registry.GetPackage(ctx, packageID.String())
	if err != nil || registered.Slug != request.ConfirmationSlug {
		return PurgeApprovalReceipt{}, false, ErrInvalid
	}
	blockers, err := ResolveDependencyBlockers(ctx, store, installation.PackageID)
	if err != nil || len(blockers) != 0 {
		return PurgeApprovalReceipt{}, false, ErrDependencyBlocked
	}
	manifest, err := application.loadManifest(
		ctx, scope, packageID, installation.ActiveVersion.String(),
	)
	if err != nil {
		return PurgeApprovalReceipt{}, false, err
	}
	plan, err := application.buildPlan(ctx, scope, store, manifest)
	if err != nil {
		return PurgeApprovalReceipt{}, false, err
	}
	backup, err := application.backup.Verify(
		ctx, scope, request.BackupReceiptID.String(), BackupRequirement{
			TenantID: identity.TenantID.String(), PackageID: packageID.String(),
			CatalogOperationID: request.OperationID.String(),
			OwnerGenerations:   ownerGenerations(plan), Now: now,
			MinimumExpiry: now.Add(90 * 24 * time.Hour),
		},
	)
	if err != nil || backup.State != catalogcomponentv2.BackupReady ||
		backup.ReceiptID != request.BackupReceiptID.String() ||
		backup.TenantID != identity.TenantID.String() ||
		backup.PackageID != packageID.String() ||
		backup.CatalogOperationID != request.OperationID.String() ||
		!validDigest(backup.SHA256) {
		return PurgeApprovalReceipt{}, false, ErrBackupBlocked
	}
	receiptID, err := application.nextIdentity()
	if err != nil {
		return PurgeApprovalReceipt{}, false, err
	}
	requestSHA, err := approvalRequestSHA(identity.TenantID, packageID, request)
	if err != nil {
		return PurgeApprovalReceipt{}, false, err
	}
	receipt := PurgeApprovalReceipt{
		ID: receiptID.String(), TenantID: identity.TenantID.String(),
		InstallationID: request.InstallationID.String(), PackageID: packageID.String(),
		OperationID: request.OperationID.String(), RequestedBy: request.RequestedBy.String(),
		ApprovedBy: identity.ActorID.String(), BackupReceiptID: backup.ReceiptID,
		BackupSHA256: backup.SHA256, ConfirmationSlug: request.ConfirmationSlug,
		Reason: request.Reason, PurgeIdempotencyKey: request.PurgeIdempotencyKey,
		RequestSHA256: requestSHA, ApprovalIdempotencyKey: request.ApprovalIdempotencyKey,
		StepUpAuthenticatedAt: stepUpAt, ApprovedAt: now,
		ExpiresAt: now.Add(30 * time.Minute), State: purgeApprovalAvailable,
	}
	receipt.ApprovalSHA256 = approvalReceiptSHA256(receipt)
	writer, ok := store.(PurgeApprovalStore)
	if !ok || writer == nil || !receipt.Valid() {
		return PurgeApprovalReceipt{}, false, ErrUnavailable
	}
	return writer.ApprovePurge(ctx, PurgeApprovalMutation{Receipt: receipt})
}

func validApprovalRequest(value PurgeApprovalRequest) bool {
	return validUUIDv7(value.InstallationID) &&
		validUUIDv7(value.OperationID) &&
		validUUIDv7(value.BackupReceiptID) &&
		validUUIDv7(value.RequestedBy) &&
		value.StepUpAtUnix > 0 && value.Reason != "" && len(value.Reason) <= 1024 &&
		purgeSlugPattern.MatchString(value.ConfirmationSlug) &&
		validIdempotency(value.PurgeIdempotencyKey) &&
		validIdempotency(value.ApprovalIdempotencyKey)
}

func approvalRequestSHA(
	tenantID, packageID uuid.UUID,
	request PurgeApprovalRequest,
) (string, error) {
	return approvalRequestDocumentSHA(purgeApprovalRequestDocument{
		TenantID: tenantID.String(), InstallationID: request.InstallationID.String(),
		PackageID: packageID.String(), OperationID: request.OperationID.String(),
		BackupReceiptID: request.BackupReceiptID.String(),
		RequestedBy:     request.RequestedBy.String(), Reason: request.Reason,
		ConfirmationSlug:    request.ConfirmationSlug,
		PurgeIdempotencyKey: request.PurgeIdempotencyKey,
	})
}
