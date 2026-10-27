package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Get(
	ctx context.Context,
	admission Admission,
	input publicresource.GetInput,
) (ReadResult, error) {
	if input.Validate() != nil {
		return ReadResult{}, ErrContract
	}
	actor, err := admission.Actor()
	if err != nil {
		return ReadResult{}, err
	}
	var result ReadResult
	err = application.runSnapshot(
		ctx, admission, func(scopedContext context.Context, scope platformtx.TenantScope) error {
			read, scopedErr := application.iam.Get(scopedContext, scope, actor, input)
			if scopedErr != nil {
				return scopedErr
			}
			result, scopedErr = application.presentRead(scopedContext, scope, read)
			return scopedErr
		},
	)
	if err != nil {
		return ReadResult{}, err
	}
	return result, nil
}
