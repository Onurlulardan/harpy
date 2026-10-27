package item

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) RequestPurge(ctx context.Context,
	scope platformtx.TenantScope, request PurgeRequest,
) (Snapshot, error) {
	request.Reason = strings.TrimSpace(request.Reason)
	store, identity, err := application.mutationScope(ctx, scope, request.MutationRequest,
		OperationPurge, PermissionPurge)
	if err != nil || len(request.Reason) < 3 || len([]byte(request.Reason)) > 2048 {
		if err != nil {
			return Snapshot{}, err
		}
		return Snapshot{}, ErrInvalid
	}
	stepUpAt, err := application.stepUp.Verify(ctx, scope, OperationPurge, request.ItemID)
	now := application.clock.Now()
	if err != nil || stepUpAt.IsZero() || stepUpAt.After(now) ||
		now.Sub(stepUpAt) > domain.StepUpMaximumAge {
		return Snapshot{}, ErrForbidden
	}
	mutation, err := application.newMutation(identity, request.ItemID, [16]byte{},
		request.RowVersion, "", nil, request.Reason, request.IdempotencyKey,
		OperationPurge, now)
	if err != nil {
		return Snapshot{}, err
	}
	mutation.StepUpAt = stepUpAt.UTC()
	return applyMutation(ctx, store.RequestPurge, mutation)
}
