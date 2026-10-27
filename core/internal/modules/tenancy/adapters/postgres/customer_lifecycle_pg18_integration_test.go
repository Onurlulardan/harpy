//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

const (
	lifecycleControlID = secondTenantID
	lifecycleTargetID  = thirdTenantID
	lifecyclePurgeID   = fourthTenantID
	lifecycleActorID   = "70000000-0000-4000-8000-000000000001"
	lifecycleWorkerID  = "70000000-0000-4000-8000-000000000002"
)

func TestCustomerLifecycleAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetTenantMigrationCatalog(t, ctx, database)
			prepareLifecycleOrigin(t, ctx, database, origin)
			assertLifecyclePrivileges(t, ctx, database)
			assertLifecycleReads(t, ctx, database)
			assertLifecycleMutations(t, ctx, database)
			assertLifecycleOffboardAndPurge(t, ctx, database)
			assertLifecycleReceiptsImmutable(t, ctx, database)
		})
	}
}
