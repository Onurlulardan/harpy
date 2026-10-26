package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestClaimRequiresBoundLeaseIdentity(t *testing.T) {
	plan := planFixture(t)
	coordinator, store, _ := coordinatorFixture(t, plan)
	attemptID := uuid.New()
	store.lease, _ = tenantapi.NewAttemptLease(plan.OperationID(), attemptID, 1, 2,
		time.Now().UTC().Add(time.Minute).Truncate(time.Microsecond))
	lease, err := coordinator.Claim(context.Background(), tenantapi.ClaimAttemptRequest{
		OperationID: plan.OperationID(), AttemptID: attemptID, LeaseOwner: "operator-1",
		LeaseTokenSHA256: strings.Repeat("a", 64), ExpectedRevision: 1,
		LeaseDuration: time.Minute,
	})
	if err != nil || lease.AttemptID() != attemptID {
		t.Fatalf("lease = %#v, %v", lease, err)
	}
}

func TestClaimRejectsUnboundedLease(t *testing.T) {
	coordinator, _, _ := coordinatorFixture(t, planFixture(t))
	_, err := coordinator.Claim(context.Background(), tenantapi.ClaimAttemptRequest{
		OperationID: uuid.New(), AttemptID: uuid.New(), LeaseOwner: "operator",
		LeaseTokenSHA256: strings.Repeat("a", 64), ExpectedRevision: 1,
		LeaseDuration: 6 * time.Minute,
	})
	if err == nil {
		t.Fatal("unbounded lease accepted")
	}
}
