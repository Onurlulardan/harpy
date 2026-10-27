package api

import (
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildBillingLicenseHTTP(
	pool *postgres.Pool,
	policy *tenancymodule.DeploymentPolicyRuntime,
	iamAccess billingLicenseIAMAccess,
	report httpedge.ErrorSink,
) (*billingmodule.PlatformLicenseHTTPHandler, error) {
	if pool == nil || policy == nil || iamAccess == nil || report == nil {
		return nil, fmt.Errorf(
			"billing platform license HTTP composition dependencies are required",
		)
	}
	access, err := newBillingLicenseAccessVerifier(iamAccess)
	if err != nil {
		return nil, err
	}
	application, err := billingmodule.NewPlatformLicenseRuntime(pool)
	if err != nil {
		return nil, fmt.Errorf("build billing platform license runtime: %w", err)
	}
	handler, err := billingmodule.NewPlatformLicenseHTTPHandler(
		billingmodule.PlatformLicenseHTTPDependencies{
			Access:      access,
			Environment: &billingLicenseEnvironmentReader{policy: policy},
			Application: application, Report: billingmodule.ErrorSink(report),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("build billing platform license HTTP handler: %w", err)
	}
	return handler, nil
}
