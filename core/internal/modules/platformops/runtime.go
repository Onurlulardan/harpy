package platformops

import (
	"context"
	"fmt"
	"time"

	platformopsstore "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres"
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	platformopsapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

type RuntimeDatabase interface {
	RuntimeOwnerDatabase()
}

type RuntimeDependencies struct {
	Settings        config.API
	Pool            RuntimeDatabase
	Monitor         *health.Monitor
	Access          HTTPAccessVerifier
	Environment     HTTPEnvironmentReader
	PermissionCache platformopsapi.PermissionCacheOperations
	Anchors         platformopsapi.RecoveryAnchors
	Report          func(error)
}

type Runtime struct {
	handler  *HTTPHandler
	observer *runtimeTraceObserver
}

type runtimeClock struct{}

func (runtimeClock) Now() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }

func NewRuntime(dependencies RuntimeDependencies) (*Runtime, error) {
	if dependencies.Pool == nil || dependencies.Monitor == nil ||
		dependencies.Access == nil || dependencies.Environment == nil ||
		dependencies.PermissionCache == nil || dependencies.Anchors == nil ||
		dependencies.Report == nil {
		return nil, fmt.Errorf("PlatformOps runtime dependencies are required")
	}
	boundStore, err := platformopsstore.NewStore(dependencies.Pool)
	if err != nil {
		return nil, fmt.Errorf("build PlatformOps store: %w", err)
	}
	store := newRuntimeStore(boundStore)
	clock := runtimeClock{}
	protection := newRuntimeProtection(
		platformopsapp.NewRecoveryProtectionService(store, clock))
	recoveryService, err := newRuntimeRecoveryService(dependencies.Settings, store, clock)
	if err != nil {
		return nil, err
	}
	recovery := newRuntimeRecovery(recoveryService)
	diagnostics := platformopsapp.NewDiagnosticReader(store, clock)
	operations := newRuntimeOperations(platformopsapp.NewOperationsService(
		&runtimeHealthReader{monitor: dependencies.Monitor, clock: clock}, store, store))
	if protection == nil || diagnostics == nil || operations == nil {
		return nil, fmt.Errorf("build PlatformOps application services")
	}
	handler, err := NewHTTPHandler(HTTPDependencies{
		Access: dependencies.Access, Environment: dependencies.Environment,
		Diagnostics: diagnostics, Protection: protection,
		Anchors:      dependencies.Anchors,
		RecoverySets: recovery, Restores: recovery,
		Operations: operations, Traces: operations, Alerts: operations,
		PermissionCache: dependencies.PermissionCache, Report: dependencies.Report,
	})
	if err != nil {
		return nil, fmt.Errorf("build PlatformOps HTTP handler: %w", err)
	}
	observer, err := newRuntimeTraceObserver(dependencies.Environment, operations)
	if err != nil {
		return nil, err
	}
	return &Runtime{handler: handler, observer: observer}, nil
}

func (runtime *Runtime) Routes() []httpedge.Route { return runtime.handler.Routes() }

func (runtime *Runtime) Observe(ctx context.Context, request httpedge.RequestObservation) error {
	return runtime.observer.Observe(ctx, request)
}
