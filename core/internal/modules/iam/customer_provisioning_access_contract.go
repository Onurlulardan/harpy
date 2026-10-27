package iam

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCustomerProvisioningAuthentication = errors.New("IAM customer provisioning authentication failed")
	ErrCustomerProvisioningForbidden      = errors.New("IAM customer provisioning access is forbidden")
	ErrCustomerProvisioningUnavailable    = errors.New("IAM customer provisioning access is unavailable")
)

type CustomerProvisioningAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	ControlTenantID     uuid.UUID
	TargetTenantID      uuid.UUID
	Capability          string
	IdempotencyKey      string
	RequestSHA256       string
}

type CustomerProvisioningAccessDecision struct {
	SubjectID, ControlTenantID uuid.UUID
	DeploymentMode, Capability string
	StepUpAt                   time.Time
	StepUpPurpose              [32]byte
}

func (decision CustomerProvisioningAccessDecision) Valid() bool {
	return decision.SubjectID != uuid.Nil && decision.ControlTenantID != uuid.Nil &&
		decision.DeploymentMode == "saas" && decision.Capability != ""
}
