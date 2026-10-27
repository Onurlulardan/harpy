package documents

import (
	"context"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) billingCommand(ctx context.Context,
	access documentshttp.Access, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	billingGroup, billing, forms, err := runtime.dependencies.Generations.
		ResolveTenantBillingGenerations(ctx, access.TenantID())
	if err != nil {
		return documentshttp.ErrUnavailable
	}
	documentsGroup, documentsGeneration, err := runtime.dependencies.Generations.
		ResolveTenantDocumentsGenerations(ctx, access.TenantID())
	values := []int64{billingGroup, billing, forms, documentsGroup, documentsGeneration}
	if err != nil || !positiveGenerations(values) {
		return documentshttp.ErrUnavailable
	}
	groups := []platformtx.CommandAuthorityGroup{
		{ID: "billing_forms", Generation: billingGroup},
		{ID: "documents", Generation: documentsGroup},
	}
	domains := []platformtx.CommandAuthorityDomain{
		{Name: "billing", Generation: billing},
		{Name: "documents", Generation: documentsGeneration},
		{Name: "forms", Generation: forms},
	}
	request, err := platformtx.AdmitMultiGroupCommandScope(access.PrincipalID(),
		access.TenantID(), operation, groups, domains, []string{"billing", "documents"})
	if err != nil {
		return documentshttp.ErrUnavailable
	}
	return runtime.dependencies.Scopes.InTenant(ctx, request, action)
}

func positiveGenerations(values []int64) bool {
	for _, value := range values {
		if value < 1 {
			return false
		}
	}
	return true
}
