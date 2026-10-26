package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ChangeState(
	ctx context.Context,
	admission Admission,
	input publicresource.StateInput,
) (MutationResult, error) {
	resource, err := input.Resource()
	if err != nil {
		return MutationResult{}, ErrContract
	}
	actor, err := admission.Actor()
	if err != nil {
		return MutationResult{}, err
	}
	operation, err := operationID(resource, "state")
	if err != nil {
		return MutationResult{}, err
	}
	var result MutationResult
	err = application.runCommand(
		ctx, admission, operation,
		func(scopedContext context.Context, scope platformtx.TenantScope) error {
			mutation, scopedErr := application.iam.ChangeState(
				scopedContext, scope, actor, input,
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
