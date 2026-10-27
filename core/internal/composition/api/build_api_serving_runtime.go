package api

import (
	"log/slog"
	"net/http"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type apiServingRuntime struct {
	public, management *http.Server
	poller             *health.Poller
}

func buildAPIServingRuntime(
	settings config.API, monitor *health.Monitor, info buildinfo.Info,
	logger *slog.Logger, admission *httpedge.Admission,
	routes http.Handler, report httpedge.ErrorSink,
	observe, usageObserve httpedge.RequestObserver, pool *postgres.Pool,
	policy *tenancymodule.DeploymentPolicyRuntime,
	moduleReadiness []func() bool,
) (apiServingRuntime, error) {
	requestObserver, err := newParallelRequestObserver(observe, usageObserve)
	if err != nil {
		return apiServingRuntime{}, err
	}
	public, management, err := buildAPIServers(
		settings, monitor, info, logger, admission, routes, report, requestObserver,
	)
	if err != nil {
		return apiServingRuntime{}, err
	}
	runtimeSettings := settings.LongRunning().Runtime()
	poller, err := health.NewPoller(
		monitor, runtimeSettings.Postgres().HealthCheckPeriod(),
		runtimeSettings.ReadinessTimeout(),
		probeReadiness(pool, policy.Ready, moduleReadiness...),
		observeReadiness(admission),
	)
	if err != nil {
		return apiServingRuntime{}, err
	}
	return apiServingRuntime{public: public, management: management, poller: poller}, nil
}
