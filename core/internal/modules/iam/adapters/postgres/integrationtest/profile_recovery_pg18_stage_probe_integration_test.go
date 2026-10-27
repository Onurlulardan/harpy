//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var errRecoveryProbeRollback = errors.New("rollback successful recovery probe")

func assertRecoveryRuntimeStages(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
	parentID uuid.UUID,
) {
	t.Helper()
	reader, err := profilebackfillpostgres.NewRecoveryReader()
	if err != nil {
		t.Fatal(err)
	}
	mutation := profileapp.MutationEvidence{
		OperationID:          uuid.MustParse("56000000-0000-4000-8000-000000000001"),
		RequestSHA256:        sha256.Sum256([]byte("recovery-stage-probe")),
		ActorID:              uuid.MustParse(publicationActorID),
		ManagePermissionID:   uuid.MustParse(publicationManagePerm),
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
	shared := profilebackfillapp.Evidence{
		OperationID: mutation.OperationID, RequestSHA256: mutation.RequestSHA256,
		ActorID: mutation.ActorID, ManagePermissionID: mutation.ManagePermissionID,
		ExpectedTokenVersion:         mutation.ExpectedTokenVersion,
		ExpectedAuthorizationVersion: mutation.ExpectedAuthorizationVersion,
	}
	stage := "scope"
	var resultState profilebackfillapp.RecoveryState
	err = scenario.runtime.coordinator.InTenant(
		ctx, recoveryCommandScope(t, ctx, scenario, "iam.profile.recover.probe"),
		func(scoped context.Context, scope platformtx.TenantScope) error {
			stage = "receipt"
			_, found, stageErr := reader.LookupReceipt(scoped, scope, shared)
			if stageErr != nil || found {
				return fmt.Errorf("receipt found=%t: %w", found, stageErr)
			}
			stage = "parent"
			contract, stageErr := reader.InspectParent(scoped, scope, parentID, shared)
			if stageErr != nil {
				return stageErr
			}
			if contract.Parent.ID != parentID {
				return fmt.Errorf("parent id = %s", contract.Parent.ID)
			}
			stage = "inventory"
			rows, stageErr := drainRecoveryInventory(
				scoped, reader, scope, parentID, shared,
			)
			if stageErr != nil || len(rows) != 3 {
				return fmt.Errorf("inventory rows=%d: %w", len(rows), stageErr)
			}
			stage = "runtime"
			result, stageErr := scenario.runtime.recoveries.Recover(
				scoped, scope, profileapp.ResourceUser, parentID, mutation,
			)
			if stageErr != nil {
				return stageErr
			}
			resultState = result.State
			return errRecoveryProbeRollback
		},
	)
	if !errors.Is(err, errRecoveryProbeRollback) ||
		resultState != profilebackfillapp.RecoveryPlanned {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) {
			t.Fatalf("recovery stage %s = %s: code=%s message=%q detail=%q "+
				"constraint=%q table=%q where=%q",
				stage, resultState, databaseError.Code, databaseError.Message,
				databaseError.Detail, databaseError.ConstraintName,
				databaseError.TableName, databaseError.Where)
		}
		t.Fatalf("recovery stage %s = %s/%v", stage, resultState, err)
	}
	var children int
	if err = scenario.db.admin.QueryRow(ctx, `SELECT count(*)
FROM t_profile_publication.iam_profile_backfill_plans
WHERE recovery_of_plan_id=$1`, parentID).Scan(&children); err != nil || children != 0 {
		t.Fatalf("rolled back recovery probe children = %d/%v", children, err)
	}
}
