package api

import (
	"context"
	"errors"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type customerProvisioningAccess struct {
	runtime *iammodule.CustomerProvisioningAccessRuntime
}

func (access customerProvisioningAccess) Verify(
	ctx context.Context,
	request tenancymodule.CustomerProvisioningAccessRequest,
) (tenancymodule.CustomerProvisioningAccessDecision, error) {
	if access.runtime == nil || request.Caller != tenancymodule.CustomerProvisioningCallerHTTP {
		return tenancymodule.CustomerProvisioningAccessDecision{}, tenancymodule.ErrCustomerProvisioningForbidden
	}
	decision, err := access.runtime.Verify(ctx,
		iammodule.CustomerProvisioningAccessRequest{
			AuthorizationValues: request.Credential.AuthorizationValues,
			TenantIDValues:      request.Credential.TenantIDValues,
			TenantSlugValues:    request.Credential.TenantSlugValues,
			ControlTenantID:     request.ControlTenantID,
			TargetTenantID:      request.TargetTenantID,
			Capability:          request.Capability, IdempotencyKey: request.IdempotencyKey,
			RequestSHA256: request.RequestSHA256,
		})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrCustomerProvisioningAuthentication):
			return tenancymodule.CustomerProvisioningAccessDecision{}, tenancymodule.ErrCustomerProvisioningAuthentication
		case errors.Is(err, iammodule.ErrCustomerProvisioningForbidden):
			return tenancymodule.CustomerProvisioningAccessDecision{}, tenancymodule.ErrCustomerProvisioningForbidden
		default:
			return tenancymodule.CustomerProvisioningAccessDecision{}, tenancymodule.ErrCustomerProvisioningUnavailable
		}
	}
	return tenancymodule.CustomerProvisioningAccessDecision{SubjectID: decision.SubjectID,
		ControlTenantID: decision.ControlTenantID,
		DeploymentMode:  tenancymodule.DeploymentMode(decision.DeploymentMode),
		Capability:      decision.Capability, StepUpAt: decision.StepUpAt,
		StepUpPurpose: decision.StepUpPurpose}, nil
}

var _ tenancymodule.CustomerProvisioningAccessVerifier = customerProvisioningAccess{}
