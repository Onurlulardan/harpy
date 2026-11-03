package environmentoperator

import (
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func provisionDiagnostic(
	plan tenantapi.OnboardingPlan,
	claim leaseClaim,
	observed tenantapi.OnboardingState,
	returned int64,
) string {
	return fmt.Sprintf(
		"operation_id=%s tenant_id=%s deployment_mode=%s onboarding_state=%s claim_revision=%d returned_revision=%d expected_state=%s",
		tenantapi.UUIDString(plan.OperationID()), tenantapi.UUIDString(plan.TenantID()),
		plan.DeploymentMode(), observed, claim.revision, returned,
		migrationNextState(plan.DeploymentMode()),
	)
}
