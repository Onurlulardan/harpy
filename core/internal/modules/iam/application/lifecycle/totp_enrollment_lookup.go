package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) TOTPEnrollmentLookup(
	ctx context.Context, scope platformtx.TenantScope, query TOTPReadQuery,
) (TOTPEnrollment, error) {
	if !service.Valid() || !tenantSnapshotUsable(ctx, scope) ||
		query.AuthenticatorID == uuid.Nil || query.ChallengeID == uuid.Nil ||
		query.ActorID != scope.ActorID() || query.PermissionID == uuid.Nil ||
		query.ExpectedTokenVersion < 1 || query.ExpectedAuthorizationVersion < 1 {
		return TOTPEnrollment{}, ErrInvalid
	}
	return service.store.TOTPEnrollmentLookup(ctx, scope, query)
}
