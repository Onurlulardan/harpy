package api

import (
	"fmt"

	auditmodule "github.com/Onurlulardan/harpy/core/internal/modules/audit"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildAuditHTTP(
	pool *platformpostgres.Pool,
	coordinator *platformtx.Coordinator,
	access auditmodule.HTTPAccessVerifier,
	report httpedge.ErrorSink,
) (*auditmodule.HTTPHandler, error) {
	application, err := auditmodule.NewApplication(auditRuntimeIDs{}, moduleRuntimeClock{})
	if err != nil {
		return nil, fmt.Errorf("build Audit application: %w", err)
	}
	runtime, err := auditmodule.NewHTTPRuntime(pool, coordinator, application)
	if err != nil {
		return nil, fmt.Errorf("build Audit runtime: %w", err)
	}
	handler, err := auditmodule.NewHTTPHandler(auditmodule.HTTPHandlerDependencies{
		Access: access, Application: runtime, Report: auditmodule.HTTPErrorSink(report),
		Correlation: auditmodule.HTTPCorrelationResolver(analyticsCorrelationResolver()),
	})
	if err != nil {
		return nil, fmt.Errorf("build Audit HTTP: %w", err)
	}
	if !handler.Ready() || len(handler.Routes()) != 5 {
		return nil, fmt.Errorf("Audit HTTP route contract is incomplete")
	}
	return handler, nil
}
