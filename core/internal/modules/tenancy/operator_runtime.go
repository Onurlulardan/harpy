package tenancy

import (
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/deploymentpolicy"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

type OperatorRuntime struct {
	DeploymentPolicy *policyapp.Initializer
	Onboarding       *onboardingapp.Coordinator
}

func NewOperatorRuntime(
	policy *policyapp.Initializer,
	onboarding *onboardingapp.Coordinator,
) (*OperatorRuntime, error) {
	if policy == nil || onboarding == nil {
		return nil, tenantapi.ErrOnboardingUnavailable
	}
	return &OperatorRuntime{
		DeploymentPolicy: policy,
		Onboarding:       onboarding,
	}, nil
}
