package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type notificationGenerations struct{ group, notifications int64 }

func (runtime *HTTPRuntime) resolve(ctx context.Context, access notificationhttp.Access,
	permission notificationhttp.Permission, operation string,
) (notificationGenerations, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || !access.ValidFor(permission, operation) {
		return notificationGenerations{}, notificationhttp.ErrForbidden
	}
	group, generation, err := runtime.resolver.ResolveTenantNotificationGenerations(
		ctx, access.TenantID())
	if err != nil || group < 1 || generation < 1 {
		return notificationGenerations{}, notificationhttp.ErrUnavailable
	}
	return notificationGenerations{group: group, notifications: generation}, nil
}

func (runtime *HTTPRuntime) command(ctx context.Context, access notificationhttp.Access,
	values notificationGenerations, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	request, err := platformtx.AdmitCommandScope(access.PrincipalID(), access.TenantID(),
		operation, "notifications", values.group,
		map[string]int64{"notifications": values.notifications})
	if err != nil {
		return notificationhttp.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, request, action)
}

func (runtime *HTTPRuntime) snapshot(ctx context.Context, access notificationhttp.Access,
	values notificationGenerations, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return notificationhttp.ErrUnavailable
	}
	authority, err := platformtx.NewSnapshotAuthorityIssuer().Issue(operation,
		map[string]int64{"notifications": values.notifications})
	if err != nil {
		return notificationhttp.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: authority, Mode: platformtx.SnapshotQuery}, action)
}
