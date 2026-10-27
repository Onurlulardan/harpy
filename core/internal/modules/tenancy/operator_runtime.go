package tenancy

import (
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/deploymentpolicy"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	migrationapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantmigration"
)

type OperatorRuntime struct {
	DeploymentPolicy *policyapp.Initializer
	Onboarding       *onboardingapp.Coordinator
	TenantMigration  *migrationapp.Service
}

func NewOperatorRuntime(
	policy *policyapp.Initializer,
	onboarding *onboardingapp.Coordinator,
	migration *migrationapp.Service,
) (*OperatorRuntime, error) {
	if policy == nil || onboarding == nil || migration == nil {
		return nil, tenantapi.ErrOnboardingUnavailable
	}
	return &OperatorRuntime{
		DeploymentPolicy: policy,
		Onboarding:       onboarding,
		TenantMigration:  migration,
	}, nil
}
