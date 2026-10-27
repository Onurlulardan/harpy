package tenancy

import (
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	workerapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerworker"
)

type CustomerWorkerClaim = workerapp.Claim
type CustomerWorkerSnapshot = workerapp.CustomerSnapshot
type CustomerWorkerFirstAdminSnapshot = workerapp.FirstAdminSnapshot
type OnboardingPlan = tenantapi.OnboardingPlan
type OnboardingPlanInput = tenantapi.OnboardingPlanInput
type OnboardingState = tenantapi.OnboardingState

const (
	OnboardingAuthorityAcknowledged = tenantapi.OnboardingAuthorityAcknowledged
	OnboardingSaaSCustomerProvision = tenantapi.OnboardingSaaSCustomerProvision
	TenantBindingSaaSCustomer       = tenantapi.TenantBindingSaaSCustomer
)

var NilUUID = tenantapi.NilUUID

func NewCustomerWorker(
	claims workerapp.ClaimStore,
	authority workerapp.AuthorityAcknowledger,
	recorder workerapp.AuthorityRecorder,
	finalizer workerapp.Finalizer,
	leaseOwner string,
) (*workerapp.Worker, error) {
	return workerapp.New(claims, authority, recorder, finalizer, leaseOwner)
}

func NewOnboardingPlan(input OnboardingPlanInput) (OnboardingPlan, error) {
	return tenantapi.NewOnboardingPlan(input)
}

func OnboardingSchemaName(tenantID UUID) (string, error) {
	return tenantapi.OnboardingSchemaName(tenantID)
}
