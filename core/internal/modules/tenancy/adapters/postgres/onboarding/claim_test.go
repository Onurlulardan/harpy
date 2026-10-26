package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func TestClaimUsesDatabaseLeaseRoutine(t *testing.T) {
	operationID, attemptID := uuid.New(), uuid.New()
	until := time.Now().UTC().Add(time.Minute).Truncate(time.Microsecond)
	storage := &postgresStore{execute: noopExecute, queryRow: func(
		_ context.Context, statement string, arguments ...any,
	) Row {
		if !strings.Contains(statement, "platform.tenancy_onboarding_claim_v2") || len(arguments) != 6 {
			t.Fatalf("claim query = %q %#v", statement, arguments)
		}
		return rowFixture(func(destinations ...any) error {
			*destinations[0].(*uuid.UUID), *destinations[1].(*uuid.UUID) = operationID, attemptID
			*destinations[2].(*int64), *destinations[3].(*int64) = 1, 2
			*destinations[4].(*time.Time) = until
			return nil
		})
	}}
	lease, err := storage.Claim(context.Background(), tenantapi.ClaimAttemptRequest{
		OperationID: operationID, AttemptID: attemptID, LeaseOwner: "operator",
		LeaseTokenSHA256: strings.Repeat("a", 64), ExpectedRevision: 1,
		LeaseDuration: time.Minute})
	if err != nil || lease.LeaseUntil() != until {
		t.Fatalf("lease = %#v, %v", lease, err)
	}
}
