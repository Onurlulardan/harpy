package tenancy

import (
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

type OperatorActivationOwner = onboardingapp.OperatorActivationOwner
type OperatorActivationCommand = onboardingapp.OperatorActivationCommand

const (
	OperatorOwnerTenancy           = onboardingapp.OperatorOwnerTenancy
	OperatorOwnerIAM               = onboardingapp.OperatorOwnerIAM
	OperatorOwnerBilling           = onboardingapp.OperatorOwnerBilling
	OperatorOwnerSystemResources   = onboardingapp.OperatorOwnerSystemResources
	OperatorOwnerPlatformOps       = onboardingapp.OperatorOwnerPlatformOps
	OperatorOwnerPlatformAuthority = onboardingapp.OperatorOwnerPlatformAuthority
	OperatorOwnerAsync             = onboardingapp.OperatorOwnerAsync
)

const (
	OperatorActivateTenant            = onboardingapp.OperatorActivateTenant
	OperatorInitializeTenantIdentity  = onboardingapp.OperatorInitializeTenantIdentity
	OperatorAssignTenantBaseOffering  = onboardingapp.OperatorAssignTenantBaseOffering
	OperatorOpenTenantUsage           = onboardingapp.OperatorOpenTenantUsage
	OperatorInitializeSystemBaseline  = onboardingapp.OperatorInitializeSystemBaseline
	OperatorInitializeRecoverySigning = onboardingapp.OperatorInitializeRecoverySigning
	OperatorBindInstanceOwner         = onboardingapp.OperatorBindInstanceOwner
	OperatorPinAuthorityHead          = onboardingapp.OperatorPinAuthorityHead
	OperatorCompleteActivation        = onboardingapp.OperatorCompleteActivation
)

type OperatorActivationRow = onboardingapp.OperatorActivationRow
type OperatorActivationExecutor = onboardingapp.OperatorActivationExecutor
type OperatorActivationResources = onboardingapp.OperatorActivationResources
type OperatorActivationBinding = onboardingapp.OperatorActivationBinding
type OperatorActivationCallback = onboardingapp.OperatorActivationCallback

type OperatorFinalizationRequest struct {
	Plan             tenantapi.OnboardingPlan
	LeaseTokenSHA256 string
	ExpectedRevision int64
}
