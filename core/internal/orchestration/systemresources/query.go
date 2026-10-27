package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Query(
	ctx context.Context,
	admission Admission,
	input publicresource.QueryInput,
) (QueryResult, error) {
	if input.Validate() != nil {
		return QueryResult{}, ErrContract
	}
	actor, err := admission.Actor()
	if err != nil {
		return QueryResult{}, err
	}
	var result QueryResult
	err = application.runSnapshot(
		ctx, admission, func(scopedContext context.Context, scope platformtx.TenantScope) error {
			query, scopedErr := application.iam.Query(scopedContext, scope, actor, input)
			if scopedErr != nil {
				return scopedErr
			}
			result, scopedErr = application.presentQuery(scopedContext, scope, query)
			return scopedErr
		},
	)
	if err != nil {
		return QueryResult{}, err
	}
	return result, nil
}
