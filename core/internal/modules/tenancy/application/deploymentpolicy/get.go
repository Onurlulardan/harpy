package deploymentpolicy

import (
	"context"
	"fmt"

	tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (query *Queries) GetDeploymentPolicy(
	ctx context.Context,
) (tenancyapi.DeploymentPolicy, error) {
	if query == nil || query.reader == nil || ctx == nil || ctx.Err() != nil {
		return tenancyapi.DeploymentPolicy{}, tenancyapi.ErrDeploymentPolicyUnavailable
	}
	policy, found, err := query.reader.GetDeploymentPolicy(ctx)
	if err != nil {
		return tenancyapi.DeploymentPolicy{}, fmt.Errorf(
			"%w: read persisted policy: %v",
			tenancyapi.ErrDeploymentPolicyUnavailable,
			err,
		)
	}
	if !found {
		return tenancyapi.DeploymentPolicy{}, tenancyapi.ErrDeploymentPolicyMissing
	}
	if !policy.IsValid() {
		return tenancyapi.DeploymentPolicy{}, tenancyapi.ErrDeploymentPolicyUnavailable
	}
	return policy, nil
}

var _ tenancyapi.DeploymentPolicyQuery = (*Queries)(nil)
