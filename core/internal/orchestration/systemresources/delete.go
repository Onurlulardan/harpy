package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Delete(
	ctx context.Context,
	admission Admission,
	input publicresource.DeleteInput,
) (DeleteResult, error) {
	resource, err := input.Resource()
	if err != nil {
		return DeleteResult{}, ErrContract
	}
	actor, err := admission.Actor()
	if err != nil {
		return DeleteResult{}, err
	}
	operation, err := operationID(resource, "delete")
	if err != nil {
		return DeleteResult{}, err
	}
	var result DeleteResult
	err = application.runCommand(
		ctx, admission, operation,
		func(scopedContext context.Context, scope platformtx.TenantScope) error {
			deleted, scopedErr := application.iam.Delete(
				scopedContext, scope, actor, input,
			)
			if scopedErr != nil {
				return scopedErr
			}
			catalog, scopedErr := deleted.Catalog()
			if scopedErr != nil {
				return scopedErr
			}
			presentation, scopedErr := application.resolvePresentation(
				scopedContext, scope, catalog,
			)
			if scopedErr != nil {
				return scopedErr
			}
			result, scopedErr = NewDeleteResult(deleted, presentation)
			return scopedErr
		},
	)
	if err != nil {
		return DeleteResult{}, err
	}
	return result, nil
}
