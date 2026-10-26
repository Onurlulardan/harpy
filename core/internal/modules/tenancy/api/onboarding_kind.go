package api

type OnboardingKind string

const (
	OnboardingEnvironmentInitialize OnboardingKind = "environment_initialize"
	OnboardingSaaSCustomerProvision OnboardingKind = "saas_customer_tenant_provision"
)

type TenantBindingPurpose string

const (
	TenantBindingOnPremisePrimary TenantBindingPurpose = "onpremise_primary"
	TenantBindingSaaSControl      TenantBindingPurpose = "saas_control"
	TenantBindingSaaSCustomer     TenantBindingPurpose = "saas_customer"
)

type AuthorityScopeKind string

const (
	AuthorityScopeEnvironment AuthorityScopeKind = "environment"
	AuthorityScopeTenant      AuthorityScopeKind = "tenant"
)

func validOnboardingCombination(
	kind OnboardingKind,
	mode DeploymentMode,
	purpose TenantBindingPurpose,
) bool {
	return kind == OnboardingEnvironmentInitialize &&
		mode == DeploymentModeOnPremise && purpose == TenantBindingOnPremisePrimary ||
		kind == OnboardingEnvironmentInitialize &&
			mode == DeploymentModeSaaS && purpose == TenantBindingSaaSControl ||
		kind == OnboardingSaaSCustomerProvision &&
			mode == DeploymentModeSaaS && purpose == TenantBindingSaaSCustomer
}
