package customerprovisioning

import (
	"context"
	"crypto/sha256"
	"time"

	provisioningv1 "github.com/Onurlulardan/harpy/contracts/customerprovisioning/v1"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

const (
	CapabilityCreate            = provisioningv1.CapabilityCreate
	CapabilityRead              = provisioningv1.CapabilityRead
	CapabilityAdminSetupReissue = provisioningv1.CapabilityAdminSetupReissue
)

type CredentialEnvelope struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	InternalEvidence    []byte
}

type AccessRequest struct {
	Caller          CallerKind
	Capability      string
	ControlTenantID uuid.UUID
	TargetTenantID  uuid.UUID
	RequestSHA256   string
	IdempotencyKey  string
	Credential      CredentialEnvelope
}

type AccessDecision struct {
	SubjectID       uuid.UUID
	ControlTenantID uuid.UUID
	DeploymentMode  tenantapi.DeploymentMode
	Capability      string
	StepUpAt        time.Time
	StepUpPurpose   [sha256.Size]byte
}

func (decision AccessDecision) validFor(request AccessRequest, now time.Time) bool {
	if decision.SubjectID == uuid.Nil || decision.ControlTenantID != request.ControlTenantID ||
		decision.DeploymentMode != tenantapi.DeploymentModeSaaS ||
		decision.Capability != request.Capability {
		return false
	}
	if request.Capability == CapabilityRead {
		return decision.StepUpAt.IsZero() && decision.StepUpPurpose == [32]byte{}
	}
	expected := ProvisionStepUpPurpose(request)
	if request.Capability == CapabilityAdminSetupReissue {
		expected = AdminSetupReissueStepUpPurpose(request)
	}
	return !decision.StepUpAt.IsZero() && decision.StepUpAt.Location() == time.UTC &&
		!decision.StepUpAt.After(now) && now.Sub(decision.StepUpAt) <= 10*time.Minute &&
		decision.StepUpPurpose == expected
}

func AdminSetupReissueStepUpPurpose(request AccessRequest) [sha256.Size]byte {
	result, _ := provisioningv1.AdminSetupReissueStepUpPurpose(
		request.ControlTenantID, request.TargetTenantID, request.IdempotencyKey,
	)
	return result
}

func ProvisionStepUpPurpose(request AccessRequest) [sha256.Size]byte {
	result, _ := provisioningv1.ProvisionStepUpPurpose(
		request.ControlTenantID, request.IdempotencyKey, request.RequestSHA256,
	)
	return result
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (AccessDecision, error)
}
