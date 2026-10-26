package dashboard

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type GrantACLRequest struct {
	DashboardID, PrincipalID uuid.UUID
	PrincipalKind            string
	Level                    string
	ExpectedRevision         uint64
	IdempotencyKey           string
}

func (application *Application) GrantACL(
	ctx context.Context, scope platformtx.TenantScope, request GrantACLRequest,
) (ACL, error) {
	permission, ok := aclPermission(request.Level)
	if !ok || !validUUIDv7(request.DashboardID) || request.PrincipalID == uuid.Nil ||
		!validPrincipalKind(request.PrincipalKind) ||
		!tokenPattern.MatchString(request.IdempotencyKey) {
		return ACL{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationGrantACL, PermissionUpdate)
	if err != nil {
		return ACL{}, err
	}
	return application.grantACLBound(ctx, scope, store, identity, request, permission)
}

func (application *Application) grantACLBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request GrantACLRequest, permission Permission,
) (ACL, error) {
	if err := application.authorizer.AuthorizePrincipal(ctx, scope,
		request.PrincipalKind, request.PrincipalID, permission); err != nil {
		return ACL{}, ErrForbidden
	}
	ids, err := application.nextIDs(1)
	if err != nil {
		return ACL{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return ACL{}, ErrUnavailable
	}
	requestDigest := aclDigest(OperationGrantACL, identity, request.DashboardID,
		request.PrincipalKind, request.PrincipalID, request.Level, request.ExpectedRevision)
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourceUpdate, ids[0], request.DashboardID,
		request.IdempotencyKey, requestDigest)
	if err != nil {
		return ACL{}, err
	}
	result, replay, err := store.GrantACL(ctx, ACLMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, DashboardID: decision.OwnerFactID,
		PrincipalID: request.PrincipalID, PrincipalKind: request.PrincipalKind, Level: request.Level,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: requestDigest,
		ExpectedRevision: request.ExpectedRevision, OccurredAt: now})
	if err != nil {
		return ACL{}, err
	}
	if result.DashboardID != decision.OwnerFactID || result.PrincipalID != request.PrincipalID ||
		result.PrincipalKind != request.PrincipalKind || result.Level != request.Level ||
		result.Revision == 0 || replay != decision.Replay {
		return ACL{}, ErrUnavailable
	}
	return result, nil
}

func validPrincipalKind(value string) bool { return value == "user" || value == "role" }

func aclPermission(level string) (Permission, bool) {
	switch level {
	case "read":
		return PermissionRead, true
	case "edit":
		return PermissionUpdate, true
	case "publish":
		return PermissionPublish, true
	default:
		return "", false
	}
}
