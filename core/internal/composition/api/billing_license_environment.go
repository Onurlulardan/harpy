package api

import (
	"context"
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type billingLicenseEnvironmentReader struct {
	policy *tenancymodule.DeploymentPolicyRuntime
}

func (reader *billingLicenseEnvironmentReader) Current(
	ctx context.Context,
) (billingmodule.PlatformLicenseHTTPEnvironment, error) {
	if reader == nil || reader.policy == nil || ctx == nil || ctx.Err() != nil {
		return billingmodule.PlatformLicenseHTTPEnvironment{},
			billingmodule.ErrPlatformLicenseHTTPUnavailable
	}
	target, err := reader.policy.PrimaryTarget()
	if err != nil || !target.IsValid() {
		return billingmodule.PlatformLicenseHTTPEnvironment{}, fmt.Errorf(
			"read billing license primary target: %w", err,
		)
	}
	return billingmodule.PlatformLicenseHTTPEnvironment{
		EnvironmentID: target.EnvironmentID(), TenantID: target.TenantID(),
		DeploymentMode: string(target.DeploymentMode()),
	}, nil
}

var _ billingmodule.PlatformLicenseHTTPEnvironmentReader = (*billingLicenseEnvironmentReader)(nil)
