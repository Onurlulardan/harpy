package systemresourcelayouts

import (
	"context"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Resolve(
	ctx context.Context,
	admission publicresource.Admission,
	resource publicresource.Resource,
) (presentation.Document, error) {
	kind, err := layoutResource(resource)
	actor, actorErr := admission.Actor()
	if err != nil || actorErr != nil {
		return presentation.Document{}, ErrContract
	}
	var result presentation.Document
	err = application.runSnapshot(
		ctx, admission,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			catalog, resolveErr := application.iam.Catalog(scoped, scope, actor, resource)
			if resolveErr != nil {
				return resolveErr
			}
			fields, resolveErr := catalog.Presentation()
			if resolveErr != nil {
				return resolveErr
			}
			result, resolveErr = application.layout.ResolvePresentation(
				scoped, scope, kind, fields,
			)
			return resolveErr
		},
	)
	return result, err
}
