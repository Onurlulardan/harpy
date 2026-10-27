package systemresourcelayout

import (
	"context"
	"time"

	layoutpostgres "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/adapters/postgres"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type layoutIDSource func() (uuid.UUID, error)

func (source layoutIDSource) NewUUID() (uuid.UUID, error) { return source() }

type LayoutRuntime struct {
	ids   layoutIDSource
	clock presentationClock
}

func NewLayoutRuntime(
	ids func() (uuid.UUID, error),
	clock func() time.Time,
) (*LayoutRuntime, error) {
	if ids == nil || clock == nil {
		return nil, layoutapi.ErrInvalid
	}
	return &LayoutRuntime{ids: layoutIDSource(ids), clock: presentationClock(clock)}, nil
}

func NewProductionLayoutRuntime(clock func() time.Time) (*LayoutRuntime, error) {
	return NewLayoutRuntime(uuid.NewV7, clock)
}

func (runtime *LayoutRuntime) CreateLayout(
	ctx context.Context, scope platformtx.TenantScope, input layoutapi.CreateInput,
) (layoutapi.Revision, error) {
	application, err := runtime.bind(scope)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	return application.Create(ctx, input)
}

func (runtime *LayoutRuntime) GetLayout(
	ctx context.Context, scope platformtx.TenantScope, id string,
) (layoutapi.Revision, error) {
	application, err := runtime.bind(scope)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	return application.Get(ctx, id)
}

func (runtime *LayoutRuntime) ListLayouts(
	ctx context.Context, scope platformtx.TenantScope, resource layoutapi.ResourceKind,
) ([]layoutapi.Revision, error) {
	application, err := runtime.bind(scope)
	if err != nil {
		return nil, err
	}
	return application.List(ctx, resource)
}

func (runtime *LayoutRuntime) UpdateLayout(
	ctx context.Context, scope platformtx.TenantScope, input layoutapi.UpdateInput,
) (layoutapi.Revision, error) {
	application, err := runtime.bind(scope)
	if err != nil {
		return layoutapi.Revision{}, err
	}
	return application.Update(ctx, input)
}

func (runtime *LayoutRuntime) DeleteLayout(
	ctx context.Context, scope platformtx.TenantScope, input layoutapi.DeleteInput,
) error {
	application, err := runtime.bind(scope)
	if err != nil {
		return err
	}
	return application.Delete(ctx, input)
}

func (runtime *LayoutRuntime) bind(
	scope platformtx.TenantScope,
) (layoutapi.LayoutAPI, error) {
	if runtime == nil || runtime.ids == nil || runtime.clock == nil {
		return nil, layoutapi.ErrInvalid
	}
	store, err := layoutpostgres.BindStore(scope)
	if err != nil {
		return nil, err
	}
	return layoutapp.NewLayoutApplication(store, runtime.ids, runtime.clock)
}

var _ layoutapi.ScopedLayout = (*LayoutRuntime)(nil)
