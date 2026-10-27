package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) demoSnapshot(
	ctx context.Context,
	access cataloghttp.Access,
	permission cataloghttp.Permission,
	operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	if runtime == nil || !access.ValidFor(permission, operation) {
		return cataloghttp.ErrForbidden
	}
	_, _, formsGeneration, err := runtime.resolver.ResolveTenantBillingGenerations(
		ctx, access.TenantID(),
	)
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	_, catalogGeneration, err := runtime.resolver.ResolveTenantCatalogGenerations(
		ctx, access.TenantID(),
	)
	if err != nil || formsGeneration < 1 || catalogGeneration < 1 {
		return cataloghttp.ErrUnavailable
	}
	actor, err := actorcontext.Issue(access.PrincipalID(), access.TenantID())
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	verified, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		operation, map[string]int64{
			"catalog": catalogGeneration,
			"forms":   formsGeneration,
		},
	)
	if err != nil {
		return cataloghttp.ErrUnavailable
	}
	return runtime.scopes.InTenant(ctx, platformtx.ScopeRequest{
		Actor: actor, Operation: verified, Mode: platformtx.SnapshotQuery,
	}, runtime.authorized(access, action))
}
