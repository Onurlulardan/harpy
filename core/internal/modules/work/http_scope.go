package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type workGenerations struct {
	group, work                  int64
	billingGroup, billing, forms int64
}

func (runtime *PlatformHTTPRuntime) resolve(ctx context.Context,
	access HTTPAccess,
) (workGenerations, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil ||
		access.PrincipalID() == uuid.Nil || access.TenantID() == uuid.Nil {
		return workGenerations{}, ErrHTTPForbidden
	}
	group, generation, err := runtime.resolver.ResolveTenantWorkGenerations(
		ctx, access.TenantID())
	if err != nil || group < 1 || generation < 1 {
		return workGenerations{}, ErrHTTPUnavailable
	}
	return workGenerations{group: group, work: generation}, nil
}

func (runtime *PlatformHTTPRuntime) resolveForms(ctx context.Context,
	access HTTPAccess, values workGenerations,
) (workGenerations, error) {
	group, billing, forms, err := runtime.resolver.ResolveTenantBillingGenerations(
		ctx, access.TenantID())
	if err != nil || group < 1 || billing < 1 || forms < 1 ||
		values.group < 1 || values.work < 1 {
		return workGenerations{}, ErrHTTPUnavailable
	}
	values.billingGroup, values.billing, values.forms = group, billing, forms
	return values, nil
}

func (runtime *PlatformHTTPRuntime) command(ctx context.Context, access HTTPAccess,
	values workGenerations, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	request, err := platformtx.AdmitCommandScope(access.PrincipalID(), access.TenantID(),
		operation, "work", values.group, map[string]int64{"work": values.work})
	if err != nil {
		return ErrHTTPUnavailable
	}
	return runtime.scopes.InTenant(ctx, request, action)
}

func (runtime *PlatformHTTPRuntime) commandForms(ctx context.Context, access HTTPAccess,
	values workGenerations, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	request, err := platformtx.AdmitMultiGroupCommandScope(
		access.PrincipalID(), access.TenantID(), operation,
		[]platformtx.CommandAuthorityGroup{
			{ID: "work", Generation: values.group},
			{ID: "billing_forms", Generation: values.billingGroup},
		},
		[]platformtx.CommandAuthorityDomain{
			{Name: "work", Generation: values.work},
			{Name: "billing", Generation: values.billing},
			{Name: "forms", Generation: values.forms},
		},
		[]string{"work", "forms"},
	)
	if err != nil {
		return ErrHTTPUnavailable
	}
	return runtime.scopes.InTenant(ctx, request, action)
}

func (runtime *PlatformHTTPRuntime) snapshot(ctx context.Context, access HTTPAccess,
	values workGenerations, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return ErrHTTPUnavailable
	}
	authority, err := platformtx.NewSnapshotAuthorityIssuer().Issue(operation,
		map[string]int64{"work": values.work})
	if err != nil {
		return ErrHTTPUnavailable
	}
	return runtime.scopes.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: authority, Mode: platformtx.SnapshotQuery}, action)
}

func (runtime *PlatformHTTPRuntime) snapshotForms(ctx context.Context, access HTTPAccess,
	values workGenerations, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return ErrHTTPUnavailable
	}
	authority, err := platformtx.NewSnapshotAuthorityIssuer().Issue(operation,
		map[string]int64{"forms": values.forms, "work": values.work})
	if err != nil {
		return ErrHTTPUnavailable
	}
	return runtime.scopes.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: authority, Mode: platformtx.SnapshotQuery}, action)
}

func (runtime *PlatformHTTPRuntime) compatibilitySnapshot(
	ctx context.Context,
	access HTTPAccess,
	values workGenerations,
	operation string,
	includeForms bool,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	if includeForms {
		return runtime.snapshotForms(ctx, access, values, operation, action)
	}
	return runtime.snapshot(ctx, access, values, operation, action)
}
