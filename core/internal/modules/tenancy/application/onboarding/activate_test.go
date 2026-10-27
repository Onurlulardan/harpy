package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	activation "github.com/Onurlulardan/harpy/core/internal/capabilities/tenantactivation/api"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestActivateRunsExactManifestInOneUnitAndClosesScope(t *testing.T) {
	plan := planFixture(t)
	coordinator, store, transaction := coordinatorFixture(t, plan)
	now := time.Now().UTC().Truncate(time.Microsecond)
	transaction.claim = tenantapi.ClaimedActivation{OperationID: plan.OperationID(),
		TenantID: plan.TenantID(), Kind: plan.Kind(), StepPlanSHA256: plan.StepPlanSHA256(),
		MigrationReceiptID: uuid.New(), MigrationReceiptSHA256: strings.Repeat("a", 64),
		Revision: 4, DatabaseNow: now, LeaseUntil: now.Add(time.Minute)}
	var captured activation.Scope
	err := coordinator.Activate(context.Background(), ActivationRequest{Plan: plan,
		OperationID: plan.OperationID(), LeaseTokenSHA256: strings.Repeat("b", 64),
		ExpectedRevision: 3}, ActivationCompletion{OperationID: plan.OperationID(),
		ReceiptID: uuid.New(), ReceiptSHA256: strings.Repeat("c", 64),
		EvidenceSHA256: strings.Repeat("d", 64)},
		func(_ context.Context, scope activation.Scope, resources ActivationResources) error {
			if resources.BindOwner == nil {
				return errors.New("activation owner binding is missing")
			}
			if _, err := resources.BindOwner(ActivationOwnerIAM); err != nil {
				return err
			}
			captured = scope
			commands := activation.Manifest(string(plan.Kind()))
			for _, command := range commands[:len(commands)-1] {
				if err := scope.Consume(command); err != nil {
					return err
				}
			}
			return nil
		})
	if err != nil || transaction.complete != 1 || store.complete != 0 ||
		transaction.boundSchema != plan.SchemaName() ||
		len(transaction.boundOwners) != 1 || transaction.boundOwners[0] != ActivationOwnerIAM ||
		captured.OperationID() != uuid.Nil {
		t.Fatalf("activation result complete=%d scope=%s err=%v",
			transaction.complete, captured.OperationID(), err)
	}
}

func TestActivateRollbackDoesNotCompleteAndInvalidatesScope(t *testing.T) {
	plan := planFixture(t)
	coordinator, store, transaction := coordinatorFixture(t, plan)
	now := time.Now().UTC().Truncate(time.Microsecond)
	transaction.claim = tenantapi.ClaimedActivation{OperationID: plan.OperationID(),
		TenantID: plan.TenantID(), Kind: plan.Kind(), StepPlanSHA256: plan.StepPlanSHA256(),
		MigrationReceiptID: uuid.New(), MigrationReceiptSHA256: strings.Repeat("a", 64),
		Revision: 4, DatabaseNow: now, LeaseUntil: now.Add(time.Minute)}
	crash := errors.New("owner command crashed")
	var captured activation.Scope
	err := coordinator.Activate(context.Background(), ActivationRequest{Plan: plan,
		OperationID: plan.OperationID(), LeaseTokenSHA256: strings.Repeat("b", 64),
		ExpectedRevision: 3}, ActivationCompletion{OperationID: plan.OperationID(),
		ReceiptID: uuid.New(), ReceiptSHA256: strings.Repeat("c", 64),
		EvidenceSHA256: strings.Repeat("d", 64)},
		func(_ context.Context, scope activation.Scope, _ ActivationResources) error {
			captured = scope
			return crash
		})
	if !errors.Is(err, crash) || transaction.complete != 0 || store.complete != 0 ||
		captured.OperationID() != uuid.Nil {
		t.Fatalf("rollback result complete=%d scope=%s err=%v",
			transaction.complete, captured.OperationID(), err)
	}
}
