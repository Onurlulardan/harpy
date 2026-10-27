package environmentoperator

import (
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type RecoverySigningResult struct {
	Generation        int64
	TrustBundleSHA256 string
	Reused            bool
}

type InitializeRequest struct {
	Manifest        installationv1.Manifest
	RecoverySigning RecoverySigningResult
}

type InitializeResult struct {
	Outcome     Outcome
	OperationID uuid.UUID
	TenantID    uuid.UUID
	FirstUserID uuid.UUID
}

type AuthorityRequest struct {
	OperationID             uuid.UUID
	EnvironmentID           uuid.UUID
	TenantID                uuid.UUID
	FirstUserID             uuid.UUID
	RootOrganizationID      uuid.UUID
	ProtectedAdminBindingID uuid.UUID
	InstanceOwnerBindingID  uuid.UUID
	RegistryBindingID       uuid.UUID
	DeploymentMode          tenantapi.DeploymentMode
	RecoverySigning         tenantapi.RecoverySigningPin
	IssuedAt                time.Time
}

type FinalizationRequest struct {
	Manifest         installationv1.Manifest
	Plan             tenantapi.OnboardingPlan
	AttemptID        uuid.UUID
	LeaseTokenSHA256 string
	ExpectedRevision int64
	LeaseUntil       time.Time
}

type preparedPlan struct {
	plan             tenantapi.OnboardingPlan
	authorityRequest AuthorityRequest
}

type leaseClaim struct {
	attemptID   uuid.UUID
	tokenSHA256 string
	revision    int64
	leaseUntil  time.Time
}
