package onboarding

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestAuthoritySynchronizerCallsOwnerRoutinesAtomically(t *testing.T) {
	synchronizer, err := NewAuthoritySynchronizer(func(
		_ context.Context,
		statement string,
		arguments ...any,
	) Row {
		if len(arguments) != 2 ||
			arguments[0] != operationID ||
			arguments[1] != manifestSHA256 ||
			!strings.Contains(statement,
				"platform.tenancy_authority_reconcile_current_v1(") ||
			strings.Contains(statement,
				"platform.billing_reconcile_current_global_authority_v1()") ||
			strings.Contains(statement, "WITH billing_global_result") {
			t.Fatalf("authority reconciliation query = %q %#v", statement, arguments)
		}
		return rowFixture(func(destinations ...any) error {
			*destinations[0].(*int) = 1
			*destinations[1].(*int) = 0
			*destinations[2].(*int) = 2
			*destinations[3].(*int) = 0
			*destinations[4].(*int) = 3
			*destinations[5].(*int) = 0
			*destinations[6].(*int) = 4
			*destinations[7].(*int) = 0
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := synchronizer.Synchronize(
		context.Background(), operationID, manifestSHA256,
	)
	if err != nil || result.BillingGlobalReconciled != 1 ||
		result.BillingTenantReconciled != 2 || result.WorkflowReconciled != 3 ||
		result.WorkflowRemaining != 0 || result.AnalyticsReconciled != 4 ||
		result.AnalyticsRemaining != 0 {
		t.Fatalf("authority reconciliation = %#v, %v", result, err)
	}
}

var (
	operationID    = uuid.MustParse("019b0208-0b4d-7c25-9a61-8021b57b3d25")
	manifestSHA256 = strings.Repeat("a", 64)
)
