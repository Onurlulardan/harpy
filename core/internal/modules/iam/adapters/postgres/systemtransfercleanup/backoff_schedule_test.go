package systemtransfercleanup

import (
	"context"
	"testing"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func TestRetryUsesExactLeaseAndBoundedSchedule(t *testing.T) {
	target := cleanupTarget()
	retryAt := cleanupNow.Add(5 * time.Minute)
	database := fixtureDatabase{query: func(statement string, arguments ...any) row {
		if statement != retrySQL || len(arguments) != 8 ||
			arguments[0] != "tenant_cleanup" || arguments[1] != target.TenantID ||
			arguments[2] != target.SourceID || arguments[3] != target.LeaseOwnerID ||
			arguments[4] != target.LeaseGeneration ||
			arguments[5] != app.SourceCleanupOutcomeUnknown ||
			arguments[6] != cleanupNow || arguments[7] != retryAt {
			t.Fatalf("retry arguments = %#v", arguments)
		}
		return fixtureRow(func(destinations ...any) error {
			*destinations[0].(*bool) = false
			return nil
		})
	}}
	replay, err := boundStore(t, database).ScheduleSourceCleanupRetry(
		context.Background(), app.SourceCleanupRetry{
			Target: target, ErrorCode: app.SourceCleanupOutcomeUnknown,
			ObservedAt: cleanupNow, RetryAt: retryAt,
		},
	)
	if err != nil || replay {
		t.Fatalf("ScheduleSourceCleanupRetry() = %v, %v", replay, err)
	}
}
