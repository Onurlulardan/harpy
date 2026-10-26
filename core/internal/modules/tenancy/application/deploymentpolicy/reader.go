package deploymentpolicy

import (
	"context"

	tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type PolicyReader interface {
	GetDeploymentPolicy(
		context.Context,
	) (tenancyapi.DeploymentPolicy, bool, error)
}
