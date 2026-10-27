package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *profileHTTPApplication) authorize(
	ctx context.Context,
	scope platformtx.TenantScope,
	claims accesstoken.Claims,
	resource profileapp.ResourceKind,
	action string,
) (uuid.UUID, error) {
	resourceKey, valid := profileResourceKey(resource)
	if !valid {
		return uuid.Nil, ErrProfileHTTPInvalid
	}
	request, err := authorization.NewRequest(
		resourceKey, action, claims.TokenVersion(), claims.AuthorizationVersion(),
	)
	if err != nil {
		return uuid.Nil, ErrProfileHTTPForbidden
	}
	decision, err := application.authorizer.Authorize(ctx, scope, request)
	if err != nil || decision.Validate() != nil {
		return uuid.Nil, ErrProfileHTTPUnavailable
	}
	if !decision.Allowed() {
		return uuid.Nil, ErrProfileHTTPForbidden
	}
	permissionID, found := decision.PermissionID()
	if !found || permissionID == uuid.Nil {
		return uuid.Nil, ErrProfileHTTPUnavailable
	}
	return permissionID, nil
}

func profileResourceKey(resource profileapp.ResourceKind) (string, bool) {
	switch resource {
	case profileapp.ResourceUser:
		return "iam.users", true
	case profileapp.ResourceRole:
		return "iam.roles", true
	case profileapp.ResourceOrganization:
		return "iam.organizations", true
	default:
		return "", false
	}
}
