package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestReconcileReturnsDatabaseEvidenceState(t *testing.T) {
	plan := planFixture(t)
	coordinator, store, _ := coordinatorFixture(t, plan)
	store.transition = tenantapi.OnboardingTransition{OperationID: plan.OperationID(),
		State: tenantapi.OnboardingMigrationReady, Revision: 5,
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond)}
	result, err := coordinator.Reconcile(context.Background(), ReconcileRequest{
		OperationID: plan.OperationID(), ReceiptID: uuid.New(),
		ReceiptSHA256: strings.Repeat("a", 64), EvidenceSHA256: strings.Repeat("b", 64),
		ExpectedRevision: 4,
	})
	if err != nil || result.State != tenantapi.OnboardingMigrationReady {
		t.Fatalf("reconcile = %#v, %v", result, err)
	}
}
