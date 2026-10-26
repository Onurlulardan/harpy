package notifications

import (
	"context"
	"reflect"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type HTTPGenerationResolver interface {
	ResolveTenantNotificationGenerations(context.Context, uuid.UUID) (int64, int64, error)
}

type HTTPScopeCoordinator interface {
	InTenant(context.Context, platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error) error
}

type HTTPRuntime struct {
	resolver    HTTPGenerationResolver
	coordinator HTTPScopeCoordinator
	application *application.Application
	realtime    *application.RealtimeApplication
	projection  ports.RealtimeProjectionBinder
}

func NewHTTPRuntime(resolver HTTPGenerationResolver, coordinator HTTPScopeCoordinator,
	base *application.Application, realtime *application.RealtimeApplication,
) (*HTTPRuntime, error) {
	if httpNil(resolver) || httpNil(coordinator) || base == nil || realtime == nil {
		return nil, notificationhttp.ErrInvalid
	}
	return &HTTPRuntime{resolver: resolver, coordinator: coordinator,
		application: base, realtime: realtime}, nil
}

func httpNil(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

var _ notificationhttp.Application = (*HTTPRuntime)(nil)
