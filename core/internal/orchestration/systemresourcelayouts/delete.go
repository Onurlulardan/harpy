package systemresourcelayouts

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Delete(
	ctx context.Context,
	admission publicresource.Admission,
	resource publicresource.Resource,
	input layoutapi.DeleteInput,
) error {
	kind, err := layoutResource(resource)
	if err != nil || input.ID == "" || input.ExpectedRowRevision < 1 {
		return ErrContract
	}
	return application.runCommand(
		ctx, admission, "system_resource_layout.delete",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			current, currentErr := application.layout.GetLayout(scoped, scope, input.ID)
			if currentErr != nil {
				return currentErr
			}
			if current.Resource != kind {
				return layoutapi.ErrNotFound
			}
			return application.layout.DeleteLayout(scoped, scope, input)
		},
	)
}
