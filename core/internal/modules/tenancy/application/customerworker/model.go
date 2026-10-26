package customerworker

import (
	"time"

	authorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type FirstAdminSnapshot struct {
	DisplayName string
	Email       string
}

type CustomerSnapshot struct {
	Name       string
	Slug       string
	FirstAdmin FirstAdminSnapshot
}

type ClaimRequest struct {
	AttemptID        tenantapi.UUID
	LeaseOwner       string
	LeaseTokenSHA256 string
	LeaseDuration    time.Duration
}

type Claim struct {
	Plan             tenantapi.OnboardingPlan
	Customer         CustomerSnapshot
	State            tenantapi.OnboardingState
	Revision         int64
	LeaseTokenSHA256 string
	LeaseUntil       time.Time
}

func (claim Claim) Valid() bool {
	return claim.Plan.IsValid() &&
		claim.Plan.Kind() == tenantapi.OnboardingSaaSCustomerProvision &&
		(claim.State == tenantapi.OnboardingAuthorityPending ||
			claim.State == tenantapi.OnboardingAuthorityAcknowledged) &&
		claim.Revision > 0 && len(claim.LeaseTokenSHA256) == 64 &&
		!claim.LeaseUntil.IsZero()
}

type AuthorityReceipt = authorityapi.AuthorityReceipt
