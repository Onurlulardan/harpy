package api

import (
	"context"
	"fmt"

	platformopsmodule "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type platformOpsEnvironmentReader struct {
	policy *tenancymodule.DeploymentPolicyRuntime
}

func (reader *platformOpsEnvironmentReader) Current(
	ctx context.Context,
) (platformopsmodule.HTTPEnvironment, error) {
	if reader == nil || reader.policy == nil || ctx == nil || ctx.Err() != nil {
		return platformopsmodule.HTTPEnvironment{}, platformopsmodule.ErrHTTPUnavailable
	}
	target, err := reader.policy.PrimaryTarget()
	if err != nil || !target.IsValid() {
		return platformopsmodule.HTTPEnvironment{}, fmt.Errorf(
			"read PlatformOps primary target: %w", err)
	}
	return platformopsmodule.HTTPEnvironment{EnvironmentID: target.EnvironmentID(),
		TenantID: target.TenantID(), DeploymentMode: string(target.DeploymentMode())}, nil
}

var _ platformopsmodule.HTTPEnvironmentReader = (*platformOpsEnvironmentReader)(nil)
