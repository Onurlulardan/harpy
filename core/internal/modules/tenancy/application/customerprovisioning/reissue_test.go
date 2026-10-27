package customerprovisioning

import (
	"context"
	"errors"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestInitialAdminSetupReissueAdmitsOnlyTerminalDeliveryFailure(t *testing.T) {
	for _, deliveryState := range []string{"permanent_failure", "outcome_unknown"} {
		t.Run(deliveryState, func(t *testing.T) {
			coordinator, access, store := provisionFixture(t, tenantapi.DeploymentModeSaaS)
			seedReissueOperation(store, deliveryState)
			result, err := coordinator.ReissueInitialAdminSetup(
				context.Background(), validReissueCommand(CallerHTTP, store.operation.TenantID),
			)
			if err != nil || result.SetupDeliveryState != "reissued" || access.calls != 1 {
				t.Fatalf("result=%#v calls=%d err=%v", result, access.calls, err)
			}
		})
	}
	for _, deliveryState := range []string{"pending", "delivered", "accepted"} {
		t.Run("denies_"+deliveryState, func(t *testing.T) {
			coordinator, _, store := provisionFixture(t, tenantapi.DeploymentModeSaaS)
			seedReissueOperation(store, deliveryState)
			_, err := coordinator.ReissueInitialAdminSetup(
				context.Background(), validReissueCommand(CallerHTTP, store.operation.TenantID),
			)
			if !errors.Is(err, ErrConflict) {
				t.Fatalf("delivery=%s err=%v", deliveryState, err)
			}
		})
	}
}

func TestInitialAdminSetupReissueRequiresExactStepUpAndReplayKey(t *testing.T) {
	coordinator, access, store := provisionFixture(t, tenantapi.DeploymentModeSaaS)
	seedReissueOperation(store, "permanent_failure")
	access.tamperStepUp = true
	command := validReissueCommand(CallerHTTP, store.operation.TenantID)
	if _, err := coordinator.ReissueInitialAdminSetup(
		context.Background(), command,
	); !errors.Is(err, ErrForbidden) {
		t.Fatalf("tampered step-up err=%v", err)
	}
	access.tamperStepUp = false
	store.operation.SetupDeliveryState = "reissued"
	storeReissueID := stableID(store.operation.OperationID,
		"initial-admin-setup-reissue:"+command.IdempotencyKey)
	storeTargetID := storeReissueID
	store.reissueOperationID = storeTargetID
	if _, err := coordinator.ReissueInitialAdminSetup(
		context.Background(), command,
	); err != nil {
		t.Fatalf("exact replay failed: %v", err)
	}
	command.IdempotencyKey = "different-reissue"
	if _, err := coordinator.ReissueInitialAdminSetup(
		context.Background(), command,
	); !errors.Is(err, ErrConflict) {
		t.Fatalf("different replay key err=%v", err)
	}
}

func TestInitialAdminSetupReissueModeGuardPrecedesEveryCaller(t *testing.T) {
	for _, caller := range []CallerKind{CallerHTTP, CallerInternal} {
		t.Run(string(caller), func(t *testing.T) {
			coordinator, access, _ := provisionFixture(t, tenantapi.DeploymentModeOnPremise)
			_, err := coordinator.ReissueInitialAdminSetup(context.Background(),
				validReissueCommand(caller,
					uuid.MustParse("60000000-0000-4000-8000-000000000001")))
			if !errors.Is(err, ErrModeDenied) || access.calls != 0 {
				t.Fatalf("err=%v capability calls=%d", err, access.calls)
			}
		})
	}
}

func seedReissueOperation(store *storeFixture, deliveryState string) {
	store.operation = Operation{
		OperationID:          uuid.MustParse("61000000-0000-4000-8000-000000000001"),
		TenantID:             uuid.MustParse("62000000-0000-4000-8000-000000000001"),
		RequestedBySubjectID: uuid.MustParse("63000000-0000-4000-8000-000000000001"),
		State:                tenantapi.OnboardingSucceeded, SetupDeliveryState: deliveryState,
		CreatedAt: provisionNow, UpdatedAt: provisionNow,
	}
}

func validReissueCommand(caller CallerKind, tenantID uuid.UUID) ReissueCommand {
	return ReissueCommand{Caller: caller, TenantID: tenantID,
		IdempotencyKey: "reissue-admin-setup"}
}
