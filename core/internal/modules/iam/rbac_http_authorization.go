package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) authorize(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityEvidenceSource,
	resourceKey string,
	actionKey string,
) (authorization.Decision, error) {
	request, err := authorization.NewRequest(
		resourceKey, actionKey,
		evidence.TokenVersion(), evidence.AuthorizationVersion(),
	)
	if err != nil {
		return authorization.Decision{}, ErrRBACHTTPForbidden
	}
	decision, err := application.authorizer.Authorize(ctx, scope, request)
	if err != nil || decision.Validate() != nil {
		return authorization.Decision{}, ErrRBACHTTPUnavailable
	}
	return decision, nil
}

func (application *rbacHTTPApplication) authorizeAdmin(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityEvidenceSource,
	action string,
) (uuid.UUID, error) {
	decision, err := application.authorize(
		ctx, scope, evidence, "iam.permissions", action,
	)
	if err != nil || !decision.Allowed() {
		return uuid.Nil, ErrRBACHTTPForbidden
	}
	permissionID, valid := decision.PermissionID()
	if !valid || permissionID == uuid.Nil {
		return uuid.Nil, ErrRBACHTTPUnavailable
	}
	return permissionID, nil
}
