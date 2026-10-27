package analytics

import (
	"net/http"
	"testing"

	publicedgepostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/publicedge"
	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	analyticsentitlement "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/entitlement"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type analyticsAggregateGenerationsFixture struct{ HTTPGenerationResolver }
type analyticsAggregateScopesFixture struct{ HTTPScopeCoordinator }
type analyticsAggregatePublicDBFixture struct{ publicedgepostgres.Database }
type analyticsAggregateLegacyFixture struct {
	compatibilityapp.PublicLegacyStore
}
type analyticsAggregatePrincipalsFixture struct{ dashboardPrincipalAuthorizer }
type analyticsAggregateEntitlementsFixture struct {
	analyticsentitlement.MutationDecider
}
type analyticsAggregateExportSchedulerFixture struct{ exportapp.Scheduler }
type analyticsAggregateExportContentFixture struct{ exportapp.ContentRuntime }
type analyticsAggregateAccessFixture struct{ HTTPAccessVerifier }

func TestAnalyticsPlatformRuntimeRejectsMissingExportDependencies(t *testing.T) {
	dependencies := analyticsAggregateDependencies(t)
	dependencies.ExportScheduler = nil
	if runtime, _, _, err := NewPlatformHTTPRuntime(dependencies); err == nil || runtime != nil {
		t.Fatal("Analytics platform runtime accepted a nil export scheduler")
	}
	var typedNil *analyticsAggregateExportSchedulerFixture
	dependencies.ExportScheduler = typedNil
	if runtime, _, _, err := NewPlatformHTTPRuntime(dependencies); err == nil || runtime != nil {
		t.Fatal("Analytics platform runtime accepted a typed-nil export scheduler")
	}
	dependencies = analyticsAggregateDependencies(t)
	var typedNilContent *analyticsAggregateExportContentFixture
	dependencies.ExportContent = typedNilContent
	if runtime, _, _, err := NewPlatformHTTPRuntime(dependencies); err == nil || runtime != nil {
		t.Fatal("Analytics platform runtime accepted a typed-nil export content runtime")
	}
}

func TestAnalyticsPlatformRuntimeRejectsMissingEntitlements(t *testing.T) {
	dependencies := analyticsAggregateDependencies(t)
	dependencies.Entitlements = nil
	if runtime, _, _, err := NewPlatformHTTPRuntime(dependencies); err == nil || runtime != nil {
		t.Fatal("Analytics platform runtime accepted nil entitlements")
	}
	var typedNil *analyticsAggregateEntitlementsFixture
	dependencies.Entitlements = typedNil
	if runtime, _, _, err := NewPlatformHTTPRuntime(dependencies); err == nil || runtime != nil {
		t.Fatal("Analytics platform runtime accepted typed-nil entitlements")
	}
}

func TestAnalyticsPlatformRuntimeBuildsCompleteHTTPHandler(t *testing.T) {
	application, public, legacy, err := NewPlatformHTTPRuntime(analyticsAggregateDependencies(t))
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHTTPHandler(HTTPHandlerDependencies{
		Access: &analyticsAggregateAccessFixture{}, Application: application,
		Exports: application, Public: public, Legacy: legacy, Report: func(error) {},
		Correlation: func(*http.Request) string { return "correlation" },
		ClientIP:    func(*http.Request) (string, error) { return "127.0.0.1", nil },
	})
	if err != nil || !handler.Ready() || len(handler.Routes()) != analyticsHTTPRouteCount {
		t.Fatalf("build Analytics HTTP handler: ready=%v routes=%d err=%v",
			handler != nil && handler.Ready(), len(handler.Routes()), err)
	}
}

func analyticsAggregateDependencies(t *testing.T) PlatformHTTPRuntimeDependencies {
	t.Helper()
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	return PlatformHTTPRuntimeDependencies{
		Generations: &analyticsAggregateGenerationsFixture{},
		Scopes:      &analyticsAggregateScopesFixture{}, PublicDB: &analyticsAggregatePublicDBFixture{},
		Legacy:       &analyticsAggregateLegacyFixture{},
		Principals:   &analyticsAggregatePrincipalsFixture{},
		Entitlements: &analyticsAggregateEntitlementsFixture{}, Catalogs: catalogs,
		ExportScheduler: &analyticsAggregateExportSchedulerFixture{},
		ExportContent:   &analyticsAggregateExportContentFixture{},
	}
}
