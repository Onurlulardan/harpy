package systemresourcelayout

import (
	"context"

	layoutpostgres "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/adapters/postgres"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *LayoutRuntime) ResolvePresentation(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource layoutapi.ResourceKind,
	catalog layoutapi.PresentationFieldCatalog,
) (layoutapi.PresentationDocument, error) {
	if runtime == nil || runtime.clock == nil {
		return layoutapi.PresentationDocument{}, layoutapi.ErrInvalid
	}
	store, err := layoutpostgres.BindStore(scope)
	if err != nil {
		return layoutapi.PresentationDocument{}, err
	}
	application, err := layoutapp.NewPresentationApplication(store, runtime.clock)
	if err != nil {
		return layoutapi.PresentationDocument{}, err
	}
	return application.Resolve(ctx, resource, catalog)
}

var _ layoutapi.ScopedResolver = (*LayoutRuntime)(nil)
