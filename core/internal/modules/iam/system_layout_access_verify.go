package iam

import (
	"context"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *SystemLayoutAccessRuntime) Verify(
	ctx context.Context,
	request SystemLayoutAccessRequest,
) (SystemLayoutAccessDecision, error) {
	if runtime == nil || runtime.security == nil || !runtime.security.Valid() ||
		ctx == nil || ctx.Err() != nil || !request.Resource.Valid() {
		return SystemLayoutAccessDecision{}, ErrSystemLayoutAccessUnavailable
	}
	target, valid := resolveSystemLayoutAccess(request.Action, request.Operation)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.clock().UTC().Truncate(time.Second)
	if !valid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return SystemLayoutAccessDecision{}, ErrSystemLayoutAccessAuthentication
	}
	claims, err := runtime.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!tenantAccessHeadersMatch(
			ctx, runtime.pool, request.TenantIDValues,
			request.TenantSlugValues, claims.Tenant(),
		) {
		return SystemLayoutAccessDecision{}, ErrSystemLayoutAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return SystemLayoutAccessDecision{}, ErrSystemLayoutAccessForbidden
	}
	if err = runtime.authorizeSystemLayout(ctx, claims, request.Resource, target); err != nil {
		return SystemLayoutAccessDecision{}, err
	}
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || len(generations) != 2 || generations["iam"] <= 0 ||
		generations["iam"] != generations["system_resource_layout"] {
		return SystemLayoutAccessDecision{}, ErrSystemLayoutAccessUnavailable
	}
	admission, err := systemLayoutAdmission(
		publicresource.ID(claims.Tenant()), publicresource.ID(claims.Subject()),
		claims.TokenVersion(), claims.AuthorizationVersion(), generations["iam"],
	)
	if err != nil {
		return SystemLayoutAccessDecision{}, ErrSystemLayoutAccessUnavailable
	}
	decision, err := NewSystemLayoutAccessDecision(admission)
	if err != nil {
		return SystemLayoutAccessDecision{}, ErrSystemLayoutAccessUnavailable
	}
	return decision, nil
}
