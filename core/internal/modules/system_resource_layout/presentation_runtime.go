package systemresourcelayout

import (
	"context"
	"time"

	layoutpostgres "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/adapters/postgres"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type presentationClock func() time.Time

func (clock presentationClock) Now() time.Time {
	return clock().UTC().Truncate(time.Microsecond)
}

type PresentationRuntime struct {
	clock presentationClock
}

func NewPresentationRuntime(
	clock func() time.Time,
) (*PresentationRuntime, error) {
	if clock == nil {
		return nil, layoutapi.ErrInvalid
	}
	return &PresentationRuntime{clock: presentationClock(clock)}, nil
}

func (runtime *PresentationRuntime) ResolvePresentation(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource layoutapi.ResourceKind,
	catalog layoutapi.PresentationFieldCatalog,
) (layoutapi.PresentationDocument, error) {
	application, err := runtime.bind(scope)
	if err != nil {
		return layoutapi.PresentationDocument{}, err
	}
	return application.Resolve(ctx, resource, catalog)
}

func (runtime *PresentationRuntime) PublishPresentation(
	ctx context.Context,
	scope platformtx.TenantScope,
	input layoutapi.PublishPresentationInput,
) (layoutapi.Revision, error) {
	application, err := runtime.bind(scope)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	return application.Publish(ctx, input)
}

func (runtime *PresentationRuntime) bind(
	scope platformtx.TenantScope,
) (*layoutapp.PresentationApplication, error) {
	if runtime == nil || runtime.clock == nil {
		return nil, layoutapi.ErrInvalid
	}
	store, err := layoutpostgres.BindStore(scope)
	if err != nil {
		return nil, err
	}
	return layoutapp.NewPresentationApplication(store, runtime.clock)
}

var _ layoutapi.ScopedResolver = (*PresentationRuntime)(nil)
var _ layoutapi.ScopedPublisher = (*PresentationRuntime)(nil)
