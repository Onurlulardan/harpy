package systemresourcelayouts

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Create(
	ctx context.Context,
	admission publicresource.Admission,
	resource publicresource.Resource,
	revisionID string,
	definition layoutapi.Definition,
) (layoutapi.Revision, error) {
	kind, err := layoutResource(resource)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	if revisionID == "" {
		return layoutapi.Revision{}, ErrContract
	}
	var result layoutapi.Revision
	err = application.runCommand(
		ctx, admission, "system_resource_layout.create",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var createErr error
			result, createErr = application.layout.CreateLayout(
				scoped, scope, layoutapi.CreateInput{
					ID: revisionID, Resource: kind, Definition: definition,
				},
			)
			return createErr
		},
	)
	return result, err
}
