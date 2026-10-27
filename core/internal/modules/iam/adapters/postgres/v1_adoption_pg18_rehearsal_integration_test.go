//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestIAMV1AdoptionRehearsalAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "v1_adoption_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMPermissionInvitation(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			seedIAMFoundationBase(t, ctx, database)
			prepareIAMV1AdoptionCutoverStub(t, ctx, database)
			applyIAMV1Adoption(t, ctx, database, origin)
			for index, tenant := range iamFoundationTenants {
				operationID := fmt.Sprintf(
					"d1000000-0000-4000-8000-00000000000%d", index+1,
				)
				runID := fmt.Sprintf(
					"d2000000-0000-4000-8000-00000000000%d", index+1,
				)
				registerIAMV1AdoptionPlan(t, ctx, database, tenant, operationID, runID)
				organizationID := applyIAMV1AdoptionBatches(
					t, ctx, database, tenant, operationID, runID,
				)
				rollbackSHA := rollbackProjectAndVerifyIAMV1Adoption(
					t, ctx, database, tenant, operationID, runID,
				)
				assertIAMV1AdoptionRollback(
					t, ctx, database, tenant, operationID, runID,
				)
				rebuiltSHA := projectAndVerifyIAMV1Adoption(
					t, ctx, database, tenant, operationID, runID,
				)
				if rebuiltSHA != rollbackSHA {
					t.Fatalf("IAM V1 rebuilt checksum %s differs from rollback rehearsal %s",
						rebuiltSHA, rollbackSHA)
				}
				assertIAMV1AdoptionResult(
					t, ctx, database, tenant, operationID, runID, organizationID,
				)
			}
			assertIAMV1AdoptionIsolation(t, ctx, database)
			assertIAMV1AdoptionPrivileges(t, ctx, database)
		})
	}
}
