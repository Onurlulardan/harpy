package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestRecordReturnsPersistedTransition(t *testing.T) {
	plan := planFixture(t)
	coordinator, store, _ := coordinatorFixture(t, plan)
	now := time.Now().UTC().Truncate(time.Microsecond)
	store.transition = tenantapi.OnboardingTransition{OperationID: plan.OperationID(),
		State: tenantapi.OnboardingMigrationReady, Revision: 3, UpdatedAt: now}
	receipt := tenantapi.OnboardingReceipt{ReceiptID: uuid.New(),
		OperationID: plan.OperationID(), AttemptID: uuid.New(),
		Kind: tenantapi.ReceiptMigrationCompleted, SHA256: strings.Repeat("a", 64),
		EvidenceSHA256: strings.Repeat("b", 64), OccurredAt: now}
	transition, err := coordinator.Record(context.Background(), receipt, 2)
	if err != nil || transition.State != tenantapi.OnboardingMigrationReady {
		t.Fatalf("transition = %#v, %v", transition, err)
	}
}
