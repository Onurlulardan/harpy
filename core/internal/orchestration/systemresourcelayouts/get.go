package systemresourcelayouts

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Get(
	ctx context.Context,
	admission publicresource.Admission,
	resource publicresource.Resource,
	id string,
) (layoutapi.Revision, error) {
	kind, err := layoutResource(resource)
	if err != nil || id == "" {
		return layoutapi.Revision{}, ErrContract
	}
	var result layoutapi.Revision
	err = application.runSnapshot(
		ctx, admission,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var getErr error
			result, getErr = application.layout.GetLayout(scoped, scope, id)
			if getErr == nil && result.Resource != kind {
				return layoutapi.ErrNotFound
			}
			return getErr
		},
	)
	return result, err
}
