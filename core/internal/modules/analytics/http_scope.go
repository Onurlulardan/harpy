package analytics

import (
	"context"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	authority "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/authority"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) command(
	ctx context.Context, access analyticshttp.Access, permission analyticshttp.Permission,
	operation string, action func(context.Context, platformtx.TenantScope) error,
) error {
	group, generation, err := runtime.resolve(ctx, access, permission, operation)
	if err != nil {
		return err
	}
	request, err := platformtx.AdmitCommandScope(access.PrincipalID(), access.TenantID(),
		operation, "analytics", group, map[string]int64{"analytics": generation})
	if err != nil {
		return analyticshttp.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, request,
		runtime.authorizedAction(access, permission, operation, action))
}

func (runtime *HTTPRuntime) snapshot(
	ctx context.Context, access analyticshttp.Access, permission analyticshttp.Permission,
	operation string, action func(context.Context, platformtx.TenantScope) error,
) error {
	_, generation, err := runtime.resolve(ctx, access, permission, operation)
	if err != nil {
		return err
	}
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return analyticshttp.ErrUnavailable
	}
	verified, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		operation, map[string]int64{"analytics": generation})
	if err != nil {
		return analyticshttp.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: verified, Mode: platformtx.SnapshotQuery},
		runtime.authorizedAction(access, permission, operation, action))
}

func (runtime *HTTPRuntime) authorizedAction(
	access analyticshttp.Access, permission analyticshttp.Permission, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) func(context.Context, platformtx.TenantScope) error {
	return func(ctx context.Context, scope platformtx.TenantScope) error {
		var evidence authority.Evidence
		var err error
		if access.StepUpAuthenticatedAt().IsZero() {
			evidence, err = authority.NewEvidence(access.PrincipalID(), access.TenantID(),
				access.TokenVersion(), access.AuthorizationVersion(),
				access.AuthorizationClock(), string(permission), operation)
		} else {
			evidence, err = authority.NewStepUpEvidence(access.PrincipalID(), access.TenantID(),
				access.TokenVersion(), access.AuthorizationVersion(),
				access.AuthorizationClock(), string(permission), operation,
				access.StepUpAuthenticatedAt())
		}
		if err != nil {
			return analyticshttp.ErrUnavailable
		}
		authorized, err := authority.WithEvidence(ctx, evidence)
		if err != nil {
			return analyticshttp.ErrUnavailable
		}
		return action(authorized, scope)
	}
}

func (runtime *HTTPRuntime) queryOwnerCommand(
	ctx context.Context, access analyticshttp.Access, permission analyticshttp.Permission,
	operation string, action func(context.Context, platformtx.TenantScope) error,
) error {
	analyticsGroup, analyticsGeneration, err := runtime.resolve(
		ctx, access, permission, operation)
	if err != nil {
		return err
	}
	billingGroup, billingGeneration, err :=
		runtime.resolver.ResolveTenantBillingGenerations(ctx, access.TenantID())
	if err != nil || billingGroup < 1 || billingGeneration < 1 {
		return analyticshttp.ErrUnavailable
	}
	request, err := platformtx.AdmitMultiGroupCommandScope(
		access.PrincipalID(), access.TenantID(), operation,
		[]platformtx.CommandAuthorityGroup{
			{ID: "analytics", Generation: analyticsGroup},
			{ID: "billing", Generation: billingGroup},
		}, []platformtx.CommandAuthorityDomain{
			{Name: "analytics", Generation: analyticsGeneration},
			{Name: "billing", Generation: billingGeneration},
		}, []string{"analytics", "billing"})
	if err != nil {
		return analyticshttp.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, request,
		runtime.authorizedAction(access, permission, operation, action))
}

func (runtime *HTTPRuntime) queryOwnerSnapshot(
	ctx context.Context, access analyticshttp.Access, permission analyticshttp.Permission,
	operation string, action func(context.Context, platformtx.TenantScope) error,
) error {
	_, analyticsGeneration, err := runtime.resolve(ctx, access, permission, operation)
	if err != nil {
		return err
	}
	_, billingGeneration, err := runtime.resolver.ResolveTenantBillingGenerations(
		ctx, access.TenantID())
	if err != nil || billingGeneration < 1 {
		return analyticshttp.ErrUnavailable
	}
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return analyticshttp.ErrUnavailable
	}
	verified, err := platformtx.NewSnapshotAuthorityIssuer().Issue(operation,
		map[string]int64{"analytics": analyticsGeneration, "billing": billingGeneration})
	if err != nil {
		return analyticshttp.ErrUnavailable
	}
	return runtime.coordinator.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: verified, Mode: platformtx.SnapshotQuery},
		runtime.authorizedAction(access, permission, operation, action))
}

func (runtime *HTTPRuntime) resolve(
	ctx context.Context, access analyticshttp.Access, permission analyticshttp.Permission,
	operation string,
) (int64, int64, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || !access.ValidFor(permission, operation) {
		return 0, 0, analyticshttp.ErrForbidden
	}
	group, generation, err := runtime.resolver.ResolveTenantAnalyticsGenerations(ctx, access.TenantID())
	if err != nil || group < 1 || generation < 1 {
		return 0, 0, analyticshttp.ErrUnavailable
	}
	return group, generation, nil
}
