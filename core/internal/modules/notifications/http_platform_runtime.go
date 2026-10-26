package notifications

import (
	"crypto/rand"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type PlatformHTTPRuntimeDependencies struct {
	Generations HTTPGenerationResolver
	Scopes      HTTPScopeCoordinator
	Catalogs    platformasync.WriteCatalogs
}

func NewPlatformHTTPRuntime(dependencies PlatformHTTPRuntimeDependencies) (*HTTPRuntime, error) {
	if httpNil(dependencies.Generations) || httpNil(dependencies.Scopes) ||
		httpNil(dependencies.Catalogs) {
		return nil, application.ErrConfiguration
	}
	binder, base, err := buildProductionNotificationApplication(dependencies.Catalogs)
	if err != nil {
		return nil, err
	}
	realtime, err := application.NewRealtimeApplication(base, binder, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("build notification realtime application: %w", err)
	}
	runtime, err := NewHTTPRuntime(dependencies.Generations, dependencies.Scopes, base, realtime)
	if err != nil {
		return nil, err
	}
	runtime.projection = binder
	return runtime, nil
}
