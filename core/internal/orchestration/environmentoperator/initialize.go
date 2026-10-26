package environmentoperator

import (
	"context"
	"errors"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) Initialize(
	ctx context.Context,
	request InitializeRequest,
) (InitializeResult, error) {
	if coordinator == nil || ctx == nil || ctx.Err() != nil ||
		!validInitializeRequest(request) {
		return InitializeResult{}, ErrInvalid
	}
	mode, _, _ := onboardingMode(request.Manifest.DeploymentMode())
	operationID := stableID(request.Manifest, "operation")
	policy, err := coordinator.policy.CreateDeploymentPolicy(ctx,
		tenantapi.InitializeDeploymentPolicyRequest{
			OperationID:    operationID,
			EnvironmentID:  uuid.UUID(request.Manifest.EnvironmentID()),
			OperationKind:  tenantapi.OnboardingEnvironmentInitialize,
			DeploymentMode: mode,
		})
	if err != nil {
		if errors.Is(err, tenantapi.ErrDeploymentPolicyConflict) {
			return InitializeResult{}, ErrFenced
		}
		return InitializeResult{}, ErrResumable
	}
	recovery := recoveryPin(request)
	prepared, err := coordinator.preparePlan(
		ctx, request.Manifest, policy.Policy.CreatedAt(), recovery,
	)
	if err != nil {
		return InitializeResult{}, err
	}
	operation, err := coordinator.onboarding.Accept(ctx, prepared.plan)
	if err != nil {
		if errors.Is(err, tenantapi.ErrOnboardingConflict) {
			return InitializeResult{}, ErrFenced
		}
		return InitializeResult{}, ErrResumable
	}
	result := InitializeResult{OperationID: operationID,
		TenantID:    uuid.UUID(prepared.plan.TenantID()),
		FirstUserID: uuid.UUID(prepared.plan.FirstUserID())}
	if operation.State() == tenantapi.OnboardingSucceeded {
		if err := coordinator.verifyNoopPinMatch(request, prepared.plan); err != nil {
			return InitializeResult{}, err
		}
		result.Outcome = OutcomeNoop
		return result, nil
	}
	if err := terminalStateError(operation.State()); err != nil {
		return InitializeResult{}, err
	}
	if err := coordinator.continueOperation(ctx, request, prepared, operation); err != nil {
		return InitializeResult{}, err
	}
	result.Outcome = OutcomeCompleted
	return result, nil
}

func validInitializeRequest(request InitializeRequest) bool {
	if len(request.Manifest.CanonicalBytes()) == 0 || request.Manifest.SHA256() == "" {
		return false
	}
	switch request.Manifest.DeploymentMode() {
	case installationv1.DeploymentOnPremise:
		return request.RecoverySigning.IsValid()
	case installationv1.DeploymentSaaS:
		return request.RecoverySigning == (recoverysigningZero())
	default:
		return false
	}
}

func recoveryPin(request InitializeRequest) tenantapi.RecoverySigningPin {
	if request.Manifest.DeploymentMode() == installationv1.DeploymentSaaS {
		return tenantapi.RecoverySigningPin{}
	}
	return tenantapi.RecoverySigningPin{Present: true,
		Generation:        request.RecoverySigning.Generation,
		TrustBundleSHA256: request.RecoverySigning.TrustBundleSHA256}
}

func (coordinator *Coordinator) verifyNoopPinMatch(
	request InitializeRequest, plan tenantapi.OnboardingPlan,
) error {
	switch request.Manifest.DeploymentMode() {
	case installationv1.DeploymentOnPremise:
		pin := plan.RecoverySigning()
		if !pin.Present || pin.TrustBundleSHA256 != request.RecoverySigning.TrustBundleSHA256 ||
			pin.Generation != request.RecoverySigning.Generation {
			return ErrFenced
		}
	case installationv1.DeploymentSaaS:
		if request.RecoverySigning != (recoverysigningZero()) {
			return ErrFenced
		}
	default:
		return ErrInvalid
	}
	return nil
}

func terminalStateError(state tenantapi.OnboardingState) error {
	switch state {
	case tenantapi.OnboardingFailed:
		return ErrFailed
	case tenantapi.OnboardingFenced, tenantapi.OnboardingFinalizing:
		return ErrFenced
	default:
		return nil
	}
}
