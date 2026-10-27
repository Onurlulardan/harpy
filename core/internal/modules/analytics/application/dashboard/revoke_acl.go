package dashboard

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type RevokeACLRequest struct {
	DashboardID, PrincipalID uuid.UUID
	PrincipalKind            string
	ExpectedRevision         uint64
	IdempotencyKey           string
}

func (application *Application) RevokeACL(
	ctx context.Context, scope platformtx.TenantScope, request RevokeACLRequest,
) error {
	if !validUUIDv7(request.DashboardID) || request.PrincipalID == uuid.Nil ||
		request.ExpectedRevision == 0 || !validPrincipalKind(request.PrincipalKind) ||
		!tokenPattern.MatchString(request.IdempotencyKey) {
		return ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationRevokeACL, PermissionUpdate)
	if err != nil {
		return err
	}
	ids, err := application.nextIDs(1)
	if err != nil {
		return err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return ErrUnavailable
	}
	requestDigest := aclDigest(OperationRevokeACL, identity, request.DashboardID,
		request.PrincipalKind, request.PrincipalID, "", request.ExpectedRevision)
	_, err = store.RevokeACL(ctx, ACLMutation{OperationID: ids[0], ActorID: identity.ActorID,
		DashboardID: request.DashboardID, PrincipalID: request.PrincipalID,
		PrincipalKind: request.PrincipalKind, ExpectedRevision: request.ExpectedRevision,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestDigest, OccurredAt: now})
	return err
}
