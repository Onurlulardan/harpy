package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *identitySystemResourcePort) Query(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	input publicresource.QueryInput,
) (publicresource.QueryResult, error) {
	evidence, resource, request, err := identitySystemResourceQueryInput(
		ctx, scope, actor, input,
	)
	if err != nil {
		return publicresource.QueryResult{}, err
	}
	authority, err := newIdentitySystemResourceAuthorization(
		port.application, ctx, scope, evidence, resource,
	)
	if err != nil {
		return publicresource.QueryResult{}, err
	}
	if _, err = authority.require("read"); err != nil {
		return publicresource.QueryResult{}, err
	}
	catalog, published, err := port.resolveCatalogSnapshot(
		ctx, scope, resource, authority,
	)
	if err != nil {
		return publicresource.QueryResult{}, err
	}
	result, err := port.application.queries.RunValue(
		ctx, scope, published, request,
	)
	if err != nil {
		return publicresource.QueryResult{}, identitySystemResourceError(err)
	}
	items := result.Items()
	relations, err := port.relationProjections(
		ctx, scope, evidence, resource, items,
	)
	if err != nil {
		return publicresource.QueryResult{}, identitySystemResourceError(err)
	}
	return publicresource.NewQueryResultWithRelations(
		items, result.HasMore(), result.NextCursor(), result.ExactCount(),
		catalog, relations,
	)
}
