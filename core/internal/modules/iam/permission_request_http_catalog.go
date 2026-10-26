package iam

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *permissionRequestHTTPApplication) ListCatalogResources(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestCatalogList,
) (PermissionRequestCatalogResourcePageView, error) {
	return application.listCatalogResourcesWithEvidence(ctx, claims, input)
}

func (application *permissionRequestHTTPApplication) listCatalogResourcesWithEvidence(
	ctx context.Context,
	evidence identityEvidenceSource,
	input PermissionRequestCatalogList,
) (PermissionRequestCatalogResourcePageView, error) {
	query := permissionRequestCatalogQuery(input)
	var result permissionrequest.CatalogResourcePage
	err := application.inSnapshot(ctx, evidence, func(
		scoped context.Context, scope platformtx.TenantScope,
	) error {
		var err error
		result, err = application.runtime.catalog.GetResources(scoped, scope, query)
		return err
	})
	if err != nil {
		return PermissionRequestCatalogResourcePageView{}, err
	}
	return permissionRequestCatalogResourcePage(result), nil
}

func (application *permissionRequestHTTPApplication) ListCatalogActions(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestCatalogList,
) (PermissionRequestCatalogActionPageView, error) {
	return application.listCatalogActionsWithEvidence(ctx, claims, input)
}

func (application *permissionRequestHTTPApplication) listCatalogActionsWithEvidence(
	ctx context.Context,
	evidence identityEvidenceSource,
	input PermissionRequestCatalogList,
) (PermissionRequestCatalogActionPageView, error) {
	query := permissionRequestCatalogQuery(input)
	var result permissionrequest.CatalogActionPage
	err := application.inSnapshot(ctx, evidence, func(
		scoped context.Context, scope platformtx.TenantScope,
	) error {
		var err error
		result, err = application.runtime.catalog.ResolveActions(scoped, scope, query)
		return err
	})
	if err != nil {
		return PermissionRequestCatalogActionPageView{}, err
	}
	return permissionRequestCatalogActionPage(result), nil
}

func (application *permissionRequestHTTPApplication) ListMissingCatalogResources(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestMissingCatalogList,
) (PermissionRequestCatalogResourcePageView, error) {
	return application.listMissingCatalogResourcesWithEvidence(ctx, claims, input)
}

func (application *permissionRequestHTTPApplication) listMissingCatalogResourcesWithEvidence(
	ctx context.Context,
	evidence identityEvidenceSource,
	input PermissionRequestMissingCatalogList,
) (PermissionRequestCatalogResourcePageView, error) {
	query := permissionrequest.MissingCatalogQuery{
		CatalogQuery: permissionRequestCatalogQuery(input.PermissionRequestCatalogList),
		ActionKey:    input.ActionKey,
	}
	var result permissionrequest.CatalogResourcePage
	err := application.inSnapshot(ctx, evidence, func(
		scoped context.Context, scope platformtx.TenantScope,
	) error {
		var err error
		result, err = application.runtime.catalog.ReconcileMissingResources(scoped, scope, query)
		return err
	})
	if err != nil {
		return PermissionRequestCatalogResourcePageView{}, err
	}
	return permissionRequestCatalogResourcePage(result), nil
}

func permissionRequestCatalogQuery(
	input PermissionRequestCatalogList,
) permissionrequest.CatalogQuery {
	return permissionrequest.CatalogQuery{
		Kind: input.Kind, Search: input.Search, Limit: input.Limit, After: input.After,
	}
}

func permissionRequestCatalogResourcePage(
	page permissionrequest.CatalogResourcePage,
) PermissionRequestCatalogResourcePageView {
	view := PermissionRequestCatalogResourcePageView{
		Items:   make([]PermissionRequestCatalogResourceView, len(page.Items)),
		HasMore: page.Next != uuid.Nil,
	}
	for index, item := range page.Items {
		resource := PermissionRequestCatalogResourceView{
			ID: item.ID.String(), Name: item.DisplayName, Slug: item.Key,
			Description: item.Description, Kind: string(item.Kind),
			IsSystem: item.Kind == permissionrequest.CatalogKindSystem,
		}
		if item.Kind == permissionrequest.CatalogKindForm {
			formSlug := strings.TrimPrefix(item.Key, "form.")
			resource.FormSlug = &formSlug
		}
		view.Items[index] = resource
	}
	if page.Next != uuid.Nil {
		next := page.Next.String()
		view.NextAfterID = &next
	}
	return view
}

func permissionRequestCatalogActionPage(
	page permissionrequest.CatalogActionPage,
) PermissionRequestCatalogActionPageView {
	view := PermissionRequestCatalogActionPageView{
		Items:   make([]PermissionRequestCatalogActionView, len(page.Items)),
		HasMore: page.Next != uuid.Nil,
	}
	for index, item := range page.Items {
		view.Items[index] = PermissionRequestCatalogActionView{
			ID: item.ID.String(), Name: item.DisplayName, Slug: item.Key,
			Description: item.Description, IsSystem: false,
		}
	}
	if page.Next != uuid.Nil {
		next := page.Next.String()
		view.NextAfterID = &next
	}
	return view
}
