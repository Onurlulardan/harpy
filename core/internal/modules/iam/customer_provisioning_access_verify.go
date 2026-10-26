package iam

import (
	"context"
	"time"

	provisioningv1 "github.com/Onurlulardan/harpy/contracts/customerprovisioning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *CustomerProvisioningAccessRuntime) Verify(
	ctx context.Context,
	request CustomerProvisioningAccessRequest,
) (CustomerProvisioningAccessDecision, error) {
	if runtime == nil || runtime.base == nil || ctx == nil || ctx.Err() != nil {
		return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningUnavailable
	}
	if !customerProvisioningCapability(request.Capability) || request.ControlTenantID == [16]byte{} {
		return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningAuthentication
	}
	wire, bearerValid := billingBearer(request.AuthorizationValues)
	now := runtime.base.clock().UTC().Truncate(time.Second)
	if !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningAuthentication
	}
	claims, err := runtime.base.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) || claims.Tenant() != request.ControlTenantID ||
		!runtime.headersMatch(ctx, request, claims.Tenant()) {
		return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningForbidden
	}
	decision, err := runtime.base.authorizePlatformBilling(ctx, claims, billingPermissionTarget{
		global: true, capability: request.Capability,
	})
	if err != nil || !decision.Allowed {
		if err == ErrBillingAccessForbidden {
			return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningForbidden
		}
		return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningUnavailable
	}
	result := CustomerProvisioningAccessDecision{SubjectID: claims.Subject(),
		ControlTenantID: claims.Tenant(), DeploymentMode: decision.DeploymentMode,
		Capability: request.Capability}
	if request.Capability == provisioningv1.CapabilityCreate ||
		request.Capability == provisioningv1.CapabilityAdminSetupReissue {
		purpose, purposeErr := customerProvisioningStepUpPurpose(request)
		stepUpAt, stepUp := customerProvisioningStepUp(claims, purpose, now)
		if purposeErr != nil || !stepUp {
			return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningForbidden
		}
		result.StepUpAt, result.StepUpPurpose = stepUpAt, purpose
	}
	if !result.Valid() {
		return CustomerProvisioningAccessDecision{}, ErrCustomerProvisioningUnavailable
	}
	return result, nil
}

func customerProvisioningStepUpPurpose(
	request CustomerProvisioningAccessRequest,
) ([32]byte, error) {
	if request.Capability == provisioningv1.CapabilityCreate {
		return provisioningv1.ProvisionStepUpPurpose(
			request.ControlTenantID, request.IdempotencyKey, request.RequestSHA256,
		)
	}
	return provisioningv1.AdminSetupReissueStepUpPurpose(
		request.ControlTenantID, request.TargetTenantID, request.IdempotencyKey,
	)
}

func customerProvisioningCapability(value string) bool {
	return value == provisioningv1.CapabilityCreate || value == provisioningv1.CapabilityRead ||
		value == provisioningv1.CapabilityAdminSetupReissue
}
