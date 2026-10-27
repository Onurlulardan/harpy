package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformopsmodule "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildPlatformOpsHTTP(
	settings config.API,
	pool *platformpostgres.Pool,
	policy *tenancymodule.DeploymentPolicyRuntime,
	monitor *health.Monitor,
	iamAccess platformOpsIAMAccess,
	report httpedge.ErrorSink,
) (*platformOpsHTTPRuntime, error) {
	if pool == nil || policy == nil || monitor == nil || iamAccess == nil || report == nil {
		return nil, fmt.Errorf("PlatformOps HTTP composition dependencies are required")
	}
	access, err := newPlatformOpsAccessVerifier(iamAccess)
	if err != nil {
		return nil, err
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return nil, fmt.Errorf("build PlatformOps permission cache coordinator: %w", err)
	}
	permissionCacheRuntime, err := iammodule.NewPermissionCacheRuntime(pool, coordinator)
	if err != nil {
		return nil, fmt.Errorf("build PlatformOps permission cache runtime: %w", err)
	}
	permissionCache := &platformOpsPermissionCache{runtime: permissionCacheRuntime}
	anchors, err := buildPlatformOpsRecoveryAnchors(settings, pool, coordinator)
	if err != nil {
		return nil, err
	}
	environment := &platformOpsEnvironmentReader{policy: policy}
	runtime, err := platformopsmodule.NewRuntime(platformopsmodule.RuntimeDependencies{
		Settings: settings, Pool: pool, Monitor: monitor,
		Access: access, Environment: environment, PermissionCache: permissionCache,
		Anchors: anchors,
		Report:  func(value error) { report(value) },
	})
	if err != nil {
		return nil, fmt.Errorf("build PlatformOps runtime: %w", err)
	}
	return &platformOpsHTTPRuntime{runtime: runtime}, nil
}
