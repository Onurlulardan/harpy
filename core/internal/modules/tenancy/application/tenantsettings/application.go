package tenantsettings

import (
	"context"
	"fmt"
	"reflect"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Application struct {
	registry    Registry
	projections ProjectionReader
	mail        MailSettings
	clock       Clock
}

func New(registry Registry, projections ProjectionReader, mail MailSettings,
	clock Clock,
) (*Application, error) {
	if missing(registry) || missing(projections) || missing(mail) || missing(clock) {
		return nil, fmt.Errorf("tenant settings dependencies are required")
	}
	return &Application{registry: registry, projections: projections,
		mail: mail, clock: clock}, nil
}

func missing(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

func (application *Application) Get(ctx context.Context,
	scope platformtx.TenantScope,
) (Snapshot, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		scope.TenantID() == uuidNil {
		return Snapshot{}, ErrInvalid
	}
	registry, err := application.registry.Read(ctx, scope)
	if err != nil {
		return Snapshot{}, err
	}
	projection, err := application.projections.Read(ctx, scope)
	if err != nil {
		return Snapshot{}, err
	}
	result := Snapshot{RegistrySnapshot: registry, Projection: projection}
	if !validSnapshot(result, scope.TenantID()) {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
