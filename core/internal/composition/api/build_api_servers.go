package api

import (
	"fmt"
	"log/slog"
	"net/http"

	platformactorbridge "github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpserver"
	"github.com/Onurlulardan/harpy/core/internal/platform/management"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func buildAPIServers(
	settings config.API,
	monitor *health.Monitor,
	info buildinfo.Info,
	logger *slog.Logger,
	admission *httpedge.Admission,
	routes http.Handler,
	report httpedge.ErrorSink,
	observer httpedge.RequestObserver,
) (*http.Server, *http.Server, error) {
	publicCore, err := httpedge.NewHandler(admission, routes, report, observer)
	if err != nil {
		return nil, nil, err
	}
	publicHandler, err := platformactorbridge.RejectPublicActorAssertion(publicCore)
	if err != nil {
		return nil, nil, fmt.Errorf("build public actor assertion guard: %w", err)
	}
	managementHandler, err := management.NewHandler(monitor, info, management.ErrorSink(report))
	if err != nil {
		return nil, nil, err
	}
	errorLog, err := telemetry.StandardErrorLog(logger)
	if err != nil {
		return nil, nil, err
	}
	publicServer, err := httpserver.NewPublic(
		settings.PublicListenAddress(), publicHandler, settings.HTTP(), errorLog)
	if err != nil {
		return nil, nil, err
	}
	managementServer, err := httpserver.NewManagement(
		settings.LongRunning().ManagementListenAddress(), managementHandler, errorLog)
	if err != nil {
		return nil, nil, err
	}
	return publicServer, managementServer, nil
}
