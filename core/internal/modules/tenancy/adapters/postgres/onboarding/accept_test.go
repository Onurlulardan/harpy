package onboarding

import (
	"context"
	"strings"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func TestAcceptUsesTypedOwnerRoutineAndCanonicalPlan(t *testing.T) {
	plan := postgresPlanFixture(t)
	storage := &postgresStore{execute: noopExecute, queryRow: func(
		_ context.Context, statement string, arguments ...any,
	) Row {
		if !strings.Contains(statement, "platform.tenancy_onboarding_accept_v2") ||
			len(arguments) != 31 || arguments[6] != plan.StepPlanSHA256() {
			t.Fatalf("accept query = %q, %d args", statement, len(arguments))
		}
		return transitionRow(plan.OperationID(), tenantapi.OnboardingAccepted, 1)
	}}
	operation, err := storage.Accept(context.Background(), plan)
	if err != nil || operation.Plan().OperationID() != plan.OperationID() {
		t.Fatalf("operation = %#v, %v", operation, err)
	}
}
