package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) resolveCatalog(ctx context.Context, access cataloghttp.Access,
	permission cataloghttp.Permission, operation string,
) (int64, int64, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || !access.ValidFor(permission, operation) {
		return 0, 0, cataloghttp.ErrForbidden
	}
	group, generation, err := runtime.resolver.ResolveTenantCatalogGenerations(ctx, access.TenantID())
	if err != nil || group < 1 || generation < 1 {
		return 0, 0, cataloghttp.ErrUnavailable
	}
	return group, generation, nil
}

func (runtime *HTTPRuntime) resolveMenuGenerations(
	ctx context.Context, access cataloghttp.Access,
	permission cataloghttp.Permission, operation string,
) (map[string]int64, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil || !access.ValidFor(permission, operation) {
		return nil, cataloghttp.ErrForbidden
	}
	_, analytics, err := runtime.resolver.ResolveTenantAnalyticsGenerations(ctx, access.TenantID())
	if err != nil {
		return nil, cataloghttp.ErrUnavailable
	}
	_, catalogGeneration, err := runtime.resolver.ResolveTenantCatalogGenerations(
		ctx, access.TenantID(),
	)
	if err != nil {
		return nil, cataloghttp.ErrUnavailable
	}
	_, workflow, err := runtime.resolver.ResolveTenantWorkflowGenerations(ctx, access.TenantID())
	if err != nil || analytics < 1 || catalogGeneration < 1 || workflow < 1 {
		return nil, cataloghttp.ErrUnavailable
	}
	return map[string]int64{
		"analytics": analytics, "catalog": catalogGeneration, "workflow": workflow,
	}, nil
}

func (runtime *HTTPRuntime) resolveLifecycle(ctx context.Context, access cataloghttp.Access,
	permission cataloghttp.Permission, operation string,
) ([]platformtx.CommandAuthorityGroup, []platformtx.CommandAuthorityDomain, error) {
	if runtime == nil || !access.ValidFor(permission, operation) {
		return nil, nil, cataloghttp.ErrUnavailable
	}
	return resolveLifecycleGenerations(ctx, runtime.resolver, access.TenantID())
}

func resolveLifecycleGenerations(
	ctx context.Context,
	resolver HTTPGenerationResolver,
	tenantID uuid.UUID,
) ([]platformtx.CommandAuthorityGroup, []platformtx.CommandAuthorityDomain, error) {
	return resolveCatalogCommandGenerations(ctx, resolver, tenantID, false)
}

func resolveResumeGenerations(
	ctx context.Context,
	resolver HTTPGenerationResolver,
	tenantID uuid.UUID,
) ([]platformtx.CommandAuthorityGroup, []platformtx.CommandAuthorityDomain, error) {
	return resolveCatalogCommandGenerations(ctx, resolver, tenantID, true)
}

func resolveCatalogCommandGenerations(
	ctx context.Context,
	resolver HTTPGenerationResolver,
	tenantID uuid.UUID,
	includeDocuments bool,
) ([]platformtx.CommandAuthorityGroup, []platformtx.CommandAuthorityDomain, error) {
	if ctx == nil || ctx.Err() != nil || nilHTTPPort(resolver) || tenantID == uuid.Nil {
		return nil, nil, cataloghttp.ErrUnavailable
	}
	analyticsGroup, analytics, err := resolver.ResolveTenantAnalyticsGenerations(ctx, tenantID)
	if err != nil {
		return nil, nil, cataloghttp.ErrUnavailable
	}
	billingGroup, billing, _, err := resolver.ResolveTenantBillingGenerations(ctx, tenantID)
	if err != nil {
		return nil, nil, cataloghttp.ErrUnavailable
	}
	catalogGroup, catalogGeneration, err := resolver.ResolveTenantCatalogGenerations(ctx, tenantID)
	if err != nil {
		return nil, nil, cataloghttp.ErrUnavailable
	}
	documentsGroup, documents := int64(1), int64(1)
	if includeDocuments {
		documentsGroup, documents, err =
			resolver.ResolveTenantDocumentsGenerations(ctx, tenantID)
		if err != nil {
			return nil, nil, cataloghttp.ErrUnavailable
		}
	}
	iamGroup, iam, layout, err := resolver.ResolveTenantIAMGenerations(ctx, tenantID)
	if err != nil {
		return nil, nil, cataloghttp.ErrUnavailable
	}
	workflowGroup, workflow, err := resolver.ResolveTenantWorkflowGenerations(ctx, tenantID)
	values := []int64{analyticsGroup, analytics, billingGroup, billing,
		catalogGroup, catalogGeneration, iamGroup, iam, layout,
		workflowGroup, workflow}
	if includeDocuments {
		values = append(values, documentsGroup, documents)
	}
	if err != nil {
		return nil, nil, cataloghttp.ErrUnavailable
	}
	for _, value := range values {
		if value < 1 {
			return nil, nil, cataloghttp.ErrUnavailable
		}
	}
	groups := []platformtx.CommandAuthorityGroup{{ID: "analytics", Generation: analyticsGroup},
		{ID: "billing_forms", Generation: billingGroup}, {ID: "catalog", Generation: catalogGroup},
		{ID: "iam_system_resources", Generation: iamGroup}, {ID: "workflow", Generation: workflowGroup}}
	domains := []platformtx.CommandAuthorityDomain{{Name: "analytics", Generation: analytics},
		{Name: "billing", Generation: billing}, {Name: "catalog", Generation: catalogGeneration},
		{Name: "iam", Generation: iam},
		{Name: "system_resource_layout", Generation: layout}, {Name: "workflow", Generation: workflow}}
	if includeDocuments {
		groups = append(groups[:3], append(
			[]platformtx.CommandAuthorityGroup{{ID: "documents", Generation: documentsGroup}},
			groups[3:]...,
		)...)
		domains = append(domains[:3], append(
			[]platformtx.CommandAuthorityDomain{{Name: "documents", Generation: documents}},
			domains[3:]...,
		)...)
	}
	return groups, domains, nil
}
