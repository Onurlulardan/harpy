package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func (application *Application) PrepareCreate(
	ctx context.Context,
	input publicresource.CreateInput,
) (publicresource.PreparedCreate, error) {
	if !application.usable(ctx) || input.Validate() != nil {
		return nil, ErrContract
	}
	prepared, err := application.iam.PrepareCreate(ctx, input)
	if err != nil {
		return nil, normalizeError(err)
	}
	if prepared == nil || prepared.Contract() != publicresource.PreparedCreateContract {
		if prepared != nil {
			prepared.Destroy()
		}
		return nil, ErrUnavailable
	}
	return prepared, nil
}
