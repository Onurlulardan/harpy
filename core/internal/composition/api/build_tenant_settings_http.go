package api

import (
	"fmt"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildTenantSettingsHTTP(
	pool *platformpostgres.Pool,
	coordinator *platformtx.Coordinator,
	access tenancymodule.TenantSettingsAccessVerifier,
	report httpedge.ErrorSink,
) (*tenancymodule.TenantSettingsHTTP, error) {
	if pool == nil || coordinator == nil || access == nil || report == nil {
		return nil, fmt.Errorf("tenant settings HTTP dependencies are incomplete")
	}
	resolver, err := newTenantSettingsGenerationResolver(pool)
	if err != nil {
		return nil, err
	}
	application, err := tenancymodule.NewTenantSettingsApplication(
		tenantSettingsProjectionReader{}, tenantSettingsMailUnavailable{},
		moduleRuntimeClock{},
	)
	if err != nil {
		return nil, fmt.Errorf("build tenant settings application: %w", err)
	}
	runtime, err := tenancymodule.NewTenantSettingsHTTPRuntime(
		resolver, coordinator, application,
		platformtx.NewTenantSettingsSnapshotIssuer(),
	)
	if err != nil {
		return nil, fmt.Errorf("build tenant settings HTTP runtime: %w", err)
	}
	handler, err := tenancymodule.NewTenantSettingsHTTP(
		access, runtime, report,
	)
	if err != nil {
		return nil, fmt.Errorf("build tenant settings HTTP: %w", err)
	}
	if !handler.Ready() || len(handler.Routes()) != 3 {
		return nil, fmt.Errorf("tenant settings HTTP route contract is incomplete")
	}
	return handler, nil
}
