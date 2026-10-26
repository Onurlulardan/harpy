package deploymentpolicy

import (
	"context"
	"errors"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (initializer *Initializer) CreateDeploymentPolicy(
	ctx context.Context,
	request tenantapi.InitializeDeploymentPolicyRequest,
) (tenantapi.DeploymentPolicyInitialization, error) {
	if initializer == nil || initializer.store == nil || ctx == nil ||
		ctx.Err() != nil || !request.IsValid() {
		return tenantapi.DeploymentPolicyInitialization{},
			tenantapi.ErrDeploymentPolicyInvalid
	}
	receipt, err := initializer.store.Create(ctx, request)
	if err != nil {
		if errors.Is(err, tenantapi.ErrDeploymentPolicyConflict) {
			return tenantapi.DeploymentPolicyInitialization{}, err
		}
		return tenantapi.DeploymentPolicyInitialization{}, fmt.Errorf(
			"create tenancy deployment policy: %w", err)
	}
	if !receipt.IsValid() || receipt.OperationID != request.OperationID ||
		receipt.Policy.EnvironmentID() != tenantapi.UUIDString(request.EnvironmentID) ||
		receipt.Policy.Mode() != request.DeploymentMode {
		return tenantapi.DeploymentPolicyInitialization{},
			tenantapi.ErrDeploymentPolicyConflict
	}
	return receipt, nil
}
