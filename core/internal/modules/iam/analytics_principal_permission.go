package iam

import (
	"context"
	"errors"
	"fmt"

	analyticsprincipalpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/analyticsprincipal"
	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var ErrAnalyticsPrincipalPermission = errors.New("IAM Analytics principal permission denied")

type AnalyticsPrincipalPermissionRuntime struct {
	users authorization.Authorizer
	roles analyticsprincipalpostgres.Reader
}

func NewAnalyticsPrincipalPermissionRuntime() (*AnalyticsPrincipalPermissionRuntime, error) {
	userPort, err := authorizationpostgres.NewAnalyticsAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM Analytics user permission port: %w", err)
	}
	users, err := authorization.NewAuthorizer(userPort)
	if err != nil {
		return nil, fmt.Errorf("build IAM Analytics user authorizer: %w", err)
	}
	roles, err := analyticsprincipalpostgres.New()
	if err != nil {
		return nil, fmt.Errorf("build IAM Analytics role permission port: %w", err)
	}
	return &AnalyticsPrincipalPermissionRuntime{users: users, roles: roles}, nil
}

func (runtime *AnalyticsPrincipalPermissionRuntime) AuthorizePrincipal(
	ctx context.Context, scope platformtx.TenantScope,
	principalKind string, principalID uuid.UUID, permission string,
) error {
	if runtime == nil || runtime.roles == nil || principalID == uuid.Nil ||
		ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireOwner(scope, "analytics") != nil {
		return ErrAnalyticsPrincipalPermission
	}
	target, found := analyticsAccessTargets[permission]
	if !found {
		return ErrAnalyticsPrincipalPermission
	}
	switch principalKind {
	case "user":
		request, err := authorization.NewCurrentActorRequest(
			target.resource, target.action, principalID)
		if err != nil {
			return ErrAnalyticsPrincipalPermission
		}
		decision, err := runtime.users.CurrentActorDecision(ctx, scope, request)
		if err != nil || decision.Validate() != nil || !decision.Allowed() {
			return ErrAnalyticsPrincipalPermission
		}
		return nil
	case "role":
		allowed, err := runtime.roles.RoleAllowed(
			ctx, scope, principalID, target.resource, target.action)
		if err != nil || !allowed {
			return ErrAnalyticsPrincipalPermission
		}
		return nil
	default:
		return ErrAnalyticsPrincipalPermission
	}
}
