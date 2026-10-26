package environmentoperator

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (coordinator *Coordinator) continueOperation(
	ctx context.Context,
	request InitializeRequest,
	prepared preparedPlan,
	operation tenantapi.OnboardingOperation,
) error {
	state, revision := operation.State(), operation.Revision()
	var claim leaseClaim
	var err error
	if state == tenantapi.OnboardingAccepted || state == tenantapi.OnboardingMigrating {
		if err = coordinator.waitClaimReady(ctx, request); err != nil {
			return err
		}
		claim, err = coordinator.claim(ctx, request.Manifest, prepared.plan.OperationID(), revision)
		if err != nil {
			return err
		}
		revision, err = coordinator.migrate(ctx, prepared.plan, claim)
		if err != nil {
			return err
		}
		state = migrationNextState(prepared.plan.DeploymentMode())
	}
	if state == tenantapi.OnboardingAuthorityPending {
		if prepared.plan.DeploymentMode() != tenantapi.DeploymentModeSaaS {
			return ErrFenced
		}
		revision, err = coordinator.acknowledgeAuthority(ctx, prepared, revision)
		if err != nil {
			return err
		}
		state = tenantapi.OnboardingAuthorityAcknowledged
	}
	if state != finalizableState(prepared.plan.DeploymentMode()) {
		return ErrFenced
	}
	if claim.attemptID == [16]byte{} {
		claim, err = coordinator.claim(ctx, request.Manifest, prepared.plan.OperationID(), revision)
		if err != nil {
			return err
		}
	} else {
		claim, err = coordinator.renew(ctx, prepared.plan.OperationID(), claim, revision)
		if err != nil {
			return err
		}
	}
	return coordinator.finalize(ctx, request, prepared.plan, claim)
}

func migrationNextState(mode tenantapi.DeploymentMode) tenantapi.OnboardingState {
	if mode == tenantapi.DeploymentModeSaaS {
		return tenantapi.OnboardingAuthorityPending
	}
	return tenantapi.OnboardingMigrationReady
}

func finalizableState(mode tenantapi.DeploymentMode) tenantapi.OnboardingState {
	if mode == tenantapi.DeploymentModeSaaS {
		return tenantapi.OnboardingAuthorityAcknowledged
	}
	return tenantapi.OnboardingMigrationReady
}
