package api

import (
	"context"
	"fmt"
	"time"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type analyticsHTTPRuntime struct {
	handler     *analyticsmodule.HTTPHandler
	application *analyticsmodule.HTTPRuntime
}

func (runtime analyticsHTTPRuntime) Ready() bool {
	return runtime.handler != nil && runtime.handler.Ready() && runtime.application != nil
}

func (runtime analyticsHTTPRuntime) Routes() []httpedge.Route {
	if !runtime.Ready() {
		return nil
	}
	return analyticsRouteGroup(runtime.handler)
}

func buildAnalyticsHTTP(
	ctx context.Context,
	settings config.API,
	pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs,
	iam iamAPIRuntime,
	report httpedge.ErrorSink,
) (analyticsHTTPRuntime, error) {
	if pool == nil || iam.analyticsAccess == nil || iam.analyticsPrincipals == nil ||
		report == nil {
		return analyticsHTTPRuntime{}, fmt.Errorf("Analytics HTTP composition dependencies are required")
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return analyticsHTTPRuntime{}, fmt.Errorf("build Analytics scope coordinator: %w", err)
	}
	access, err := newAnalyticsAccessVerifier(iam.analyticsAccess)
	if err != nil {
		return analyticsHTTPRuntime{}, err
	}
	legacyStore, err := analyticsmodule.NewLegacyPublicStore(
		analyticsLegacyPublicDatabase{pool: pool},
	)
	if err != nil {
		return analyticsHTTPRuntime{}, fmt.Errorf("build Analytics legacy public store: %w", err)
	}
	entitlements, err := billingmodule.NewMeteredResourceEntitlements(
		catalogs, uuid.NewV7, time.Now,
	)
	if err != nil {
		return analyticsHTTPRuntime{}, fmt.Errorf("build Analytics billing entitlements: %w", err)
	}
	exportScheduler, err := analyticsmodule.NewDashboardExportScheduler(catalogs)
	if err != nil {
		return analyticsHTTPRuntime{}, fmt.Errorf("build Analytics export scheduler: %w", err)
	}
	exportContent, err := buildAnalyticsExportContent(ctx, settings)
	if err != nil {
		return analyticsHTTPRuntime{}, err
	}
	application, public, legacy, err := analyticsmodule.NewPlatformHTTPRuntime(
		analyticsmodule.PlatformHTTPRuntimeDependencies{
			Generations: pool, Scopes: coordinator, PublicDB: pool,
			Legacy:       legacyStore,
			Principals:   iam.analyticsPrincipals,
			Entitlements: entitlements, Catalogs: catalogs,
			ExportScheduler: exportScheduler, ExportContent: exportContent,
		})
	if err != nil {
		return analyticsHTTPRuntime{}, fmt.Errorf("build Analytics HTTP applications: %w", err)
	}
	handler, err := analyticsmodule.NewHTTPHandler(analyticsmodule.HTTPHandlerDependencies{
		Access: access, Application: application, Exports: application,
		Public: public, Legacy: legacy,
		Report:      analyticsmodule.HTTPErrorSink(report),
		Correlation: analyticsCorrelationResolver(),
		ClientIP:    newAnalyticsClientIPResolver(settings.AnalyticsTrustedProxies()),
	})
	if err != nil {
		return analyticsHTTPRuntime{}, fmt.Errorf("build Analytics HTTP handler: %w", err)
	}
	if !handler.Ready() {
		return analyticsHTTPRuntime{}, fmt.Errorf("build Analytics HTTP handler: not ready")
	}
	return analyticsHTTPRuntime{handler: handler, application: application}, nil
}

func analyticsRouteGroup(handler *analyticsmodule.HTTPHandler) []httpedge.Route {
	if handler == nil {
		return nil
	}
	routes := handler.Routes()
	result := make([]httpedge.Route, len(routes))
	for index, route := range routes {
		result[index] = httpedge.Route{
			Method: route.Method, Pattern: route.Pattern, Handler: route.Handler,
		}
	}
	return result
}
