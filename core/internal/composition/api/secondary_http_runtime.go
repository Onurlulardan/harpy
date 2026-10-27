package api

import (
	"context"
	"fmt"

	auditmodule "github.com/Onurlulardan/harpy/core/internal/modules/audit"
	supportmodule "github.com/Onurlulardan/harpy/core/internal/modules/support"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type secondaryHTTPRuntime struct {
	usage          usageHTTPRuntime
	support        *supportmodule.HTTPHandler
	audit          *auditmodule.HTTPHandler
	tenantSettings *tenancymodule.TenantSettingsHTTP
	usageObserver  httpedge.RequestObserver
}

func buildSecondaryHTTPRuntime(
	settings config.API, pool *platformpostgres.Pool,
	coordinator *platformtx.Coordinator, iam *iamAPIRuntime,
	report httpedge.ErrorSink, environmentID string,
) (secondaryHTTPRuntime, error) {
	if pool == nil || coordinator == nil || iam == nil || report == nil {
		return secondaryHTTPRuntime{}, fmt.Errorf("secondary HTTP dependencies are incomplete")
	}
	usageAccess, err := newUsageAccessVerifier(iam.usageAccess)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	supportAccess, err := newSupportAccessVerifier(iam.supportAccess)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	auditAccess, err := newAuditAccessVerifier(iam.auditAccess)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	usageHTTP, err := buildUsageHTTP(pool, usageAccess, report)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	environment, err := tenancymodule.ParseUUID(environmentID)
	if err != nil {
		return secondaryHTTPRuntime{}, fmt.Errorf("parse Usage environment identity: %w", err)
	}
	usageObserver, err := newUsageRequestObserver(environment, usageHTTP)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	supportHTTP, err := buildSupportHTTP(settings, pool, supportAccess, report)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	auditHTTP, err := buildAuditHTTP(pool, coordinator, auditAccess, report)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	tenantSettingsAccess, err := newTenantSettingsAccessVerifier(iam.tenantSettingsAccess)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	tenantSettingsHTTP, err := buildTenantSettingsHTTP(
		pool, coordinator, tenantSettingsAccess, report,
	)
	if err != nil {
		return secondaryHTTPRuntime{}, err
	}
	runtime := secondaryHTTPRuntime{usage: usageHTTP, support: supportHTTP,
		audit: auditHTTP, tenantSettings: tenantSettingsHTTP, usageObserver: usageObserver}
	if !runtime.Ready() {
		return secondaryHTTPRuntime{}, fmt.Errorf("secondary HTTP runtime is not ready")
	}
	return runtime, nil
}

func (runtime secondaryHTTPRuntime) Ready() bool {
	return runtime.usage.Ready() && runtime.usageObserver != nil &&
		runtime.support != nil && runtime.support.Ready() &&
		runtime.audit != nil && runtime.audit.Ready() &&
		runtime.tenantSettings != nil && runtime.tenantSettings.Ready()
}

func (runtime secondaryHTTPRuntime) Observe(
	ctx context.Context, observation httpedge.RequestObservation,
) error {
	if runtime.usageObserver == nil {
		return fmt.Errorf("Usage request observer is unavailable")
	}
	return runtime.usageObserver(ctx, observation)
}

func (runtime secondaryHTTPRuntime) RouteGroups() [][]httpedge.Route {
	if !runtime.Ready() {
		return nil
	}
	return [][]httpedge.Route{
		runtime.usage.Routes(), runtime.support.Routes(), runtime.audit.Routes(),
		runtime.tenantSettings.Routes(),
	}
}
