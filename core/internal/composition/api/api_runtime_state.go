package api

import (
	"net/http"

	documentsmodule "github.com/Onurlulardan/harpy/core/internal/modules/documents"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

type apiRuntime struct {
	admission            *httpedge.Admission
	poller               *health.Poller
	management           *http.Server
	public               *http.Server
	telemetryRuntime     *telemetry.Runtime
	deploymentPolicy     *tenancymodule.DeploymentPolicyRuntime
	asyncWriteCatalogs   platformasync.WriteCatalogs
	analytics            analyticsHTTPRuntime
	catalog              catalogHTTPRuntime
	documents            *documentsmodule.HTTPHandler
	documentsProduction  *documentsmodule.Production
	platformOps          *platformOpsHTTPRuntime
	notifications        notificationsHTTPRuntime
	work                 workHTTPRuntime
	secondary            secondaryHTTPRuntime
	systemLayouts        *systemResourceLayoutHTTPHandler
	customerProvisioning *customerProvisioningRuntime
	iam                  iamAPIRuntime
}
