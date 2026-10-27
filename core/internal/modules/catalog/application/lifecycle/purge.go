package lifecycle

import (
	"context"
	"encoding/hex"
	"time"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Purge(
	ctx context.Context,
	scope platformtx.TenantScope,
	request PurgeRequest,
) (OperationHandle, error) {
	if !validUUIDv7(request.OperationID) || request.PackageID == uuid.Nil ||
		!validUUIDv7(request.BackupReceiptID) || !validUUIDv7(request.ApprovalReceiptID) ||
		!validUUIDv7(request.RequestedBy) ||
		!validIdempotency(request.IdempotencyKey) ||
		request.Reason == "" || len(request.Reason) > 1024 || request.ConfirmationSlug == "" ||
		request.StepUpAtUnix <= 0 {
		return OperationHandle{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationPurge, ActionPurge)
	if err != nil {
		return OperationHandle{}, err
	}
	if request.RequestedBy != identity.ActorID {
		return OperationHandle{}, ErrForbidden
	}
	packageID, _ := packagecontract.ParsePackageID(request.PackageID.String())
	installation, err := store.GetInstallation(ctx, packageID)
	if err != nil {
		return OperationHandle{}, ErrNotFound
	}
	if installation.State == catalogdomain.InstallationPurgePending {
		return application.existingPurge(ctx, store, identity, request)
	}
	if installation.State != catalogdomain.InstallationRetired {
		return OperationHandle{}, ErrConflict
	}
	now, err := application.now()
	if err != nil || now.Sub(time.Unix(request.StepUpAtUnix, 0).UTC()) < 0 ||
		now.Sub(time.Unix(request.StepUpAtUnix, 0).UTC()) > 10*time.Minute {
		return OperationHandle{}, ErrForbidden
	}
	registry, err := application.registry.Bind(ctx, scope)
	if err != nil || registry == nil {
		return OperationHandle{}, ErrUnavailable
	}
	registered, err := registry.GetPackage(ctx, request.PackageID.String())
	if err != nil || registered.Slug != request.ConfirmationSlug {
		return OperationHandle{}, ErrInvalid
	}
	blockers, err := ResolveDependencyBlockers(ctx, store, packageID)
	if err != nil || len(blockers) != 0 {
		return OperationHandle{}, ErrDependencyBlocked
	}
	manifest, err := application.loadManifest(ctx, scope, request.PackageID, installation.ActiveVersion.String())
	if err != nil {
		return OperationHandle{}, err
	}
	plan, err := application.buildPlan(ctx, scope, store, manifest)
	if err != nil {
		return OperationHandle{}, err
	}
	operationID := request.OperationID
	backup, err := application.verifyBackup(
		ctx, scope, identity, request, operationID, plan, now,
	)
	if err != nil {
		return OperationHandle{}, err
	}
	approval, err := application.approvalForPurge(
		ctx, store, identity, installation.ID.String(), request, backup, now,
	)
	if err != nil {
		return OperationHandle{}, err
	}
	pending, err := installation.BeginPurge()
	if err != nil {
		return OperationHandle{}, ErrConflict
	}
	stepUpAt := time.Unix(request.StepUpAtUnix, 0).UTC()
	evidence, found := authorizationReceiptFromContext(ctx, scope, "catalog.purge")
	if !found {
		return OperationHandle{}, ErrForbidden
	}
	tokenVersion, authorizationVersion, authorizationClock,
		authorizationPermissionID, authorizationSHA256, stepUpPurposeSHA256 :=
		evidence.values()
	authority := PurgeAuthority{
		TenantID: identity.TenantID.String(), PackageID: request.PackageID.String(),
		PackageSlug: registered.Slug, OperationID: operationID.String(),
		ActorID: identity.ActorID.String(), RequestedBy: request.RequestedBy.String(),
		ApprovedBy: approval.ApprovedBy, ApprovalReceiptID: approval.ID,
		ApprovalSHA256:  approval.ApprovalSHA256,
		BackupReceiptID: backup.ReceiptID,
		BackupState:     backup.State, BackupSHA256: backup.SHA256,
		ConfirmationSlug: request.ConfirmationSlug, Reason: request.Reason,
		IdempotencyKey: request.IdempotencyKey, ReferenceEpoch: pending.ReferenceEpoch,
		AuthorizationTokenVersion:   tokenVersion,
		AuthorizationVersion:        authorizationVersion,
		AuthorizationClock:          authorizationClock,
		AuthorizationPermissionID:   authorizationPermissionID.String(),
		AuthorizationEvidenceSHA256: hex.EncodeToString(authorizationSHA256[:]),
		StepUpPurposeSHA256:         hex.EncodeToString(stepUpPurposeSHA256[:]),
		StepUpAuthenticatedAt:       stepUpAt, RequestedAt: now,
		ApprovalStepUpAt: approval.StepUpAuthenticatedAt,
		ApprovedAt:       approval.ApprovedAt, ApprovalExpiresAt: approval.ExpiresAt,
		CoolingUntil: now.Add(24 * time.Hour),
	}
	authority.AuthoritySHA256 = catalogcomponentv2.AuthoritySHA256(authority)
	if !authority.Valid() {
		return OperationHandle{}, ErrUnavailable
	}
	return application.start(ctx, scope, store, identity, startInput{
		Kind: catalogdomain.OperationPurge, OperationID: operationID,
		PackageID: request.PackageID, Version: installation.ActiveVersion.String(),
		IdempotencyKey: request.IdempotencyKey, Installation: pending,
		PreviousState: &installation.State, MutateInstallation: true,
		RequestDocument: authority, PurgeAuthority: &authority,
	})
}

func (application *Application) verifyBackup(
	ctx context.Context,
	scope platformtx.TenantScope,
	identity ScopeIdentity,
	request PurgeRequest,
	operationID uuid.UUID,
	plan catalogdomain.Plan,
	now time.Time,
) (BackupEvidence, error) {
	evidence, err := application.backup.Verify(
		ctx, scope, request.BackupReceiptID.String(), BackupRequirement{
			TenantID: identity.TenantID.String(), PackageID: request.PackageID.String(),
			CatalogOperationID: operationID.String(), OwnerGenerations: ownerGenerations(plan),
			Now: now, MinimumExpiry: now.Add(90 * 24 * time.Hour),
		})
	if err != nil {
		return BackupEvidence{}, ErrBackupBlocked
	}
	if evidence.ReceiptID != request.BackupReceiptID.String() ||
		evidence.TenantID != identity.TenantID.String() ||
		evidence.PackageID != request.PackageID.String() ||
		evidence.CatalogOperationID != operationID.String() ||
		evidence.State != catalogcomponentv2.BackupReady ||
		!validDigest(evidence.SHA256) {
		return BackupEvidence{}, ErrBackupBlocked
	}
	return evidence, nil
}
