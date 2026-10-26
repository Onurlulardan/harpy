package systemresourcelayouts

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) List(
	ctx context.Context,
	admission publicresource.Admission,
	resource publicresource.Resource,
) ([]layoutapi.Revision, error) {
	kind, err := layoutResource(resource)
	if err != nil {
		return nil, err
	}
	var result []layoutapi.Revision
	err = application.runSnapshot(
		ctx, admission,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var listErr error
			result, listErr = application.layout.ListLayouts(scoped, scope, kind)
			return listErr
		},
	)
	return result, err
}
