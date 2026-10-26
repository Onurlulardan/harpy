package deploymentpolicy

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type InitializationStore interface {
	Create(
		context.Context,
		tenantapi.InitializeDeploymentPolicyRequest,
	) (tenantapi.DeploymentPolicyInitialization, error)
}

type Initializer struct{ store InitializationStore }

func NewInitializer(store InitializationStore) (*Initializer, error) {
	if store == nil {
		return nil, tenantapi.ErrDeploymentPolicyUnavailable
	}
	return &Initializer{store: store}, nil
}
