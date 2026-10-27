package systemresourcelayout

import (
	"context"

	layoutpostgres "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/adapters/postgres"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *LayoutRuntime) PublishPresentation(
	ctx context.Context,
	scope platformtx.TenantScope,
	input layoutapi.PublishPresentationInput,
) (layoutapi.Revision, error) {
	if runtime == nil || runtime.clock == nil {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	store, err := layoutpostgres.BindStore(scope)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	application, err := layoutapp.NewPresentationApplication(store, runtime.clock)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	return application.Publish(ctx, input)
}

var _ layoutapi.ScopedPublisher = (*LayoutRuntime)(nil)
