package lifecycle

import (
	"context"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func (application *Application) existingPurge(
	ctx context.Context,
	store Store,
	identity ScopeIdentity,
	request PurgeRequest,
) (OperationHandle, error) {
	authorityStore, ok := store.(PurgeAuthorityStore)
	if !ok || authorityStore == nil {
		return OperationHandle{}, ErrUnavailable
	}
	operation, err := store.ReadOperation(ctx, request.OperationID)
	if err != nil {
		return OperationHandle{}, err
	}
	authority, err := authorityStore.AuthorityForPurge(ctx, request.OperationID)
	if err != nil {
		return OperationHandle{}, err
	}
	stepUp := time.Unix(request.StepUpAtUnix, 0).UTC()
	if operation.Kind != catalogdomain.OperationPurge ||
		operation.ID.String() != request.OperationID.String() ||
		operation.Plan.PackageID.String() != request.PackageID.String() ||
		operation.Plan.Slug != authority.PackageSlug || !authority.Valid() ||
		authority.PackageID != request.PackageID.String() ||
		authority.OperationID != request.OperationID.String() ||
		authority.TenantID != identity.TenantID.String() ||
		authority.ActorID != identity.ActorID.String() ||
		authority.RequestedBy != request.RequestedBy.String() ||
		authority.ApprovalReceiptID != request.ApprovalReceiptID.String() ||
		authority.BackupReceiptID != request.BackupReceiptID.String() ||
		authority.ConfirmationSlug != request.ConfirmationSlug ||
		authority.Reason != request.Reason ||
		authority.IdempotencyKey != request.IdempotencyKey ||
		!authority.StepUpAuthenticatedAt.Equal(stepUp) {
		return OperationHandle{}, ErrConflict
	}
	return OperationHandle{Operation: operation, Replayed: true}, nil
}
