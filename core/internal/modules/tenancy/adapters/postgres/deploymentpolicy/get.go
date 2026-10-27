package deploymentpolicy

import (
	"context"
	"fmt"
	"time"

	tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (reader *policyReader) GetDeploymentPolicy(
	ctx context.Context,
) (tenancyapi.DeploymentPolicy, bool, error) {
	if !reader.usable(ctx) {
		return tenancyapi.DeploymentPolicy{}, false, errPolicyUnavailable
	}
	var environmentID *string
	var mode *string
	var maximum *int64
	var createdAt *time.Time
	var found bool
	err := reader.queryRow(ctx, getDeploymentPolicySQL).Scan(
		&environmentID,
		&mode,
		&maximum,
		&createdAt,
		&found,
	)
	if err != nil {
		return tenancyapi.DeploymentPolicy{}, false,
			fmt.Errorf("read tenancy deployment policy: %w", err)
	}
	if !found {
		return tenancyapi.DeploymentPolicy{}, false, nil
	}
	if environmentID == nil || mode == nil || createdAt == nil {
		return tenancyapi.DeploymentPolicy{}, false, errPolicyUnavailable
	}
	policy, err := tenancyapi.NewDeploymentPolicy(
		*environmentID,
		tenancyapi.DeploymentMode(*mode),
		maximum,
		createdAt.UTC(),
	)
	if err != nil {
		return tenancyapi.DeploymentPolicy{}, false, errPolicyUnavailable
	}
	return policy, true, nil
}
