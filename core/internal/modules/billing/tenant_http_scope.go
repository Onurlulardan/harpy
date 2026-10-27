package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const billingAuthorityGroup = "billing_forms"

type generations struct {
	group, billing int64
}

func (application *tenantOperations) resolve(
	ctx context.Context,
	access billingapi.Access,
	permission billingapi.Permission,
) (generations, error) {
	if application == nil || application.resolver == nil || ctx == nil ||
		ctx.Err() != nil || !access.ValidFor(permission) ||
		access.Scope() != billingapi.AccessScopeTenant {
		return generations{}, billingapi.ErrForbidden
	}
	group, billing, _, err := application.resolver.ResolveTenantBillingGenerations(
		ctx, access.TenantID(),
	)
	if err != nil || group < 1 || billing < 1 {
		return generations{}, billingapi.ErrAuthorityUnavailable
	}
	return generations{group: group, billing: billing}, nil
}

func (application *tenantOperations) inCommand(
	ctx context.Context,
	access billingapi.Access,
	authority generations,
	operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	request, err := platformtx.AdmitCommandScope(
		access.PrincipalID(), access.TenantID(), operation,
		billingAuthorityGroup, authority.group,
		map[string]int64{"billing": authority.billing},
	)
	if err != nil {
		return billingapi.ErrAuthorityUnavailable
	}
	return application.coordinator.InTenant(ctx, request, action)
}

func (application *tenantOperations) inSnapshot(
	ctx context.Context,
	access billingapi.Access,
	authority generations,
	operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	actor, err := actorcontext.Issue(
		access.PrincipalID(), access.TenantID(),
	)
	if err != nil {
		return billingapi.ErrAuthorityUnavailable
	}
	verified, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		operation, map[string]int64{"billing": authority.billing},
	)
	if err != nil {
		return billingapi.ErrAuthorityUnavailable
	}
	return application.coordinator.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: verified, Mode: platformtx.SnapshotQuery,
	}, action)
}
