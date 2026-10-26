package systemresourcelayouts

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type compatibilityConfigSnapshot struct {
	presentation CompatibilityPresentation
	revisions    []layoutapi.Revision
}

func (application *Application) compatibilityConfigSnapshot(
	ctx context.Context,
	admission publicresource.Admission,
	resource publicresource.Resource,
) (compatibilityConfigSnapshot, error) {
	kind, err := layoutResource(resource)
	actor, actorErr := admission.Actor()
	if err != nil || actorErr != nil {
		return compatibilityConfigSnapshot{}, ErrContract
	}
	var result compatibilityConfigSnapshot
	err = application.runSnapshot(
		ctx, admission,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var scopedErr error
			result.presentation, scopedErr = application.compatibilityPresentationInScope(
				scoped, scope, actor, resource, kind,
			)
			if scopedErr != nil {
				return scopedErr
			}
			result.revisions, scopedErr = application.layout.ListLayouts(
				scoped, scope, kind,
			)
			return scopedErr
		},
	)
	return result, err
}
