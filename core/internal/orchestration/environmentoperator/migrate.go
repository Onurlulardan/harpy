package environmentoperator

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) migrate(
	ctx context.Context,
	plan tenantapi.OnboardingPlan,
	claim leaseClaim,
) (int64, error) {
	tenantID := uuid.UUID(plan.TenantID())
	if err := coordinator.provision.Provision(ctx, tenantID); err != nil {
		return 0, Resumable("tenant_schema.provision", "provision_failed", "", err)
	}
	return coordinator.refreshAfterProvision(ctx, plan, claim)
}

func (coordinator *Coordinator) refreshAfterProvision(
	ctx context.Context,
	plan tenantapi.OnboardingPlan,
	claim leaseClaim,
) (int64, error) {
	operation, err := coordinator.onboarding.Accept(ctx, plan)
	if err != nil {
		return 0, Resumable("tenant_schema.refresh_onboarding", "onboarding_accept_failed",
			provisionDiagnostic(plan, claim, "", 0), err)
	}
	if operation.State() != migrationNextState(plan.DeploymentMode()) ||
		operation.Revision() <= claim.revision {
		return 0, Resumable("tenant_schema.refresh_onboarding", "onboarding_state_mismatch",
			provisionDiagnostic(plan, claim, operation.State(), operation.Revision()), nil)
	}
	return operation.Revision(), nil
}
