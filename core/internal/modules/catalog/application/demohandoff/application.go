package demohandoff

import (
	"context"
	"reflect"

	api "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/demohandoff"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const owner = "catalog"

type IDSource interface {
	NewUUIDv7() (uuid.UUID, error)
}

type StoreBinder func(
	context.Context,
	platformtx.TenantScope,
	platformtx.OwnerAuthority,
) (Store, error)

type Store interface {
	Persist(context.Context, Projection) (api.Receipt, error)
}

type Application struct {
	binder StoreBinder
	ids    IDSource
}

func New(binder StoreBinder, ids IDSource) (*Application, error) {
	if nilDependency(binder) || nilDependency(ids) {
		return nil, api.ErrInvalid
	}
	return &Application{binder: binder, ids: ids}, nil
}

func nilDependency(value any) bool {
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
