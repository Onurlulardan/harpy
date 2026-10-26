package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Create(
	ctx context.Context,
	admission Admission,
	prepared publicresource.PreparedCreate,
) (MutationResult, error) {
	if prepared == nil || prepared.Contract() != publicresource.PreparedCreateContract {
		return MutationResult{}, ErrContract
	}
	resource, err := prepared.Resource()
	if err != nil {
		return MutationResult{}, ErrContract
	}
	actor, err := admission.Actor()
	if err != nil {
		return MutationResult{}, err
	}
	operation, err := operationID(resource, "create")
	if err != nil {
		return MutationResult{}, err
	}
	var result MutationResult
	err = application.runCommand(
		ctx, admission, operation,
		func(scopedContext context.Context, scope platformtx.TenantScope) error {
			mutation, scopedErr := application.iam.Create(
				scopedContext, scope, actor, prepared,
			)
			if scopedErr != nil {
				return scopedErr
			}
			result, scopedErr = application.presentMutation(scopedContext, scope, mutation)
			return scopedErr
		},
	)
	if err != nil {
		return MutationResult{}, err
	}
	return result, nil
}
