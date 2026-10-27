package tenancy

import (
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	provisioningapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/google/uuid"
)

type CustomerProvisioningAccessRequest = provisioningapp.AccessRequest
type CustomerProvisioningAccessDecision = provisioningapp.AccessDecision
type CustomerProvisioningAccessVerifier = provisioningapp.AccessVerifier
type CustomerProvisioningAuthorityRequest = provisioningapp.AuthorityRequest
type CustomerProvisioningAuthorityPlanner = provisioningapp.AuthorityPlanner
type CustomerProvisioningOfferingResolver = provisioningapp.OfferingResolver
type CustomerProvisioningSetupRequest = provisioningapp.SetupRequest
type CustomerProvisioningSetupReissueRequest = provisioningapp.SetupReissueRequest
type CustomerProvisioningSetupPreparer = provisioningapp.SetupPreparer
type CustomerProvisioningClock = provisioningapp.Clock

type BaseOfferingPin = tenantapi.BaseOfferingPin
type OfferingSourcePin = tenantapi.OfferingSourcePin
type ProtectedAuthorityPlan = tenantapi.ProtectedAuthorityPlan
type AuthorityAcknowledgement = tenantapi.AuthorityAcknowledgement
type AuthorityScopeKind = tenantapi.AuthorityScopeKind
type UUID = tenantapi.UUID

const (
	CustomerProvisioningCallerHTTP = provisioningapp.CallerHTTP
	AuthorityScopeEnvironment      = tenantapi.AuthorityScopeEnvironment
	AuthorityScopeTenant           = tenantapi.AuthorityScopeTenant
)

var (
	ErrCustomerProvisioningAuthentication = provisioningapp.ErrAuthentication
	ErrCustomerProvisioningForbidden      = provisioningapp.ErrForbidden
	ErrCustomerProvisioningUnavailable    = provisioningapp.ErrUnavailable
)

type CustomerProvisioningDependencies struct {
	Policy    provisioningapp.DeploymentPolicy
	Access    CustomerProvisioningAccessVerifier
	Offering  CustomerProvisioningOfferingResolver
	Authority CustomerProvisioningAuthorityPlanner
	Setup     CustomerProvisioningSetupPreparer
	Store     provisioningapp.OperationStore
	Clock     CustomerProvisioningClock
}

func NewCustomerProvisioningApplication(
	dependencies CustomerProvisioningDependencies,
) (CustomerProvisioningApplication, error) {
	return provisioningapp.New(provisioningapp.Dependencies{
		Policy: dependencies.Policy, Access: dependencies.Access,
		Offering: dependencies.Offering, Authority: dependencies.Authority,
		Setup: dependencies.Setup, Store: dependencies.Store, Clock: dependencies.Clock,
	})
}

func ParseUUID(value string) (UUID, error) {
	parsed, err := uuid.Parse(value)
	return UUID(parsed), err
}

func UUIDString(value UUID) string {
	return tenantapi.UUIDString(value)
}
