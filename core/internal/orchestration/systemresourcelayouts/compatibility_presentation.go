package systemresourcelayouts

import (
	"context"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CompatibilityPresentation struct {
	Document    presentation.Document
	Catalog     presentation.FieldCatalog
	Descriptors []publicresource.FieldDescriptor
}

func (application *Application) CompatibilityPresentation(
	ctx context.Context,
	admission publicresource.Admission,
	resource publicresource.Resource,
) (CompatibilityPresentation, error) {
	kind, err := layoutResource(resource)
	actor, actorErr := admission.Actor()
	if err != nil || actorErr != nil {
		return CompatibilityPresentation{}, ErrContract
	}
	var result CompatibilityPresentation
	err = application.runSnapshot(ctx, admission,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var resolveErr error
			result, resolveErr = application.compatibilityPresentationInScope(
				scoped, scope, actor, resource, kind,
			)
			return resolveErr
		})
	return result, err
}

func (application *Application) compatibilityPresentationInScope(
	ctx context.Context,
	scope platformtx.TenantScope,
	actor publicresource.ActorEvidence,
	resource publicresource.Resource,
	kind layoutapi.ResourceKind,
) (CompatibilityPresentation, error) {
	catalog, err := application.iam.Catalog(ctx, scope, actor, resource)
	if err != nil {
		return CompatibilityPresentation{}, err
	}
	fields, err := catalog.Presentation()
	if err != nil {
		return CompatibilityPresentation{}, err
	}
	descriptors, err := catalog.Descriptors()
	if err != nil {
		return CompatibilityPresentation{}, err
	}
	document, err := application.layout.ResolvePresentation(ctx, scope, kind, fields)
	if err != nil {
		return CompatibilityPresentation{}, err
	}
	return CompatibilityPresentation{
		Document: document, Catalog: fields, Descriptors: descriptors,
	}, nil
}
