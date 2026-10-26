package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) presentQuery(
	ctx context.Context,
	scope platformtx.TenantScope,
	query publicresource.QueryResult,
) (QueryResult, error) {
	catalog, err := query.Catalog()
	if err != nil {
		return QueryResult{}, ErrContract
	}
	presentation, skip, err := application.recordPresentation(ctx, scope, catalog)
	if err != nil {
		return QueryResult{}, err
	}
	if skip {
		return newIAMOnlyQueryResult(query)
	}
	return NewQueryResult(query, presentation)
}
