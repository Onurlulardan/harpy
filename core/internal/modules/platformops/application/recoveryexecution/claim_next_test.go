package recoveryexecution

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestClaimNextAcceptsOnlyCompleteDurableClaim(t *testing.T) {
	fixture := &executionStoreFixture{found: true, claim: ports.RecoveryExecutionClaim{
		AttemptID: testAttemptID, LeaseToken: testLeaseToken, AttemptGeneration: 1,
		RecoveryState: "holding", RecoverySetID: testSetID,
		OperationID:   "20000000-0000-4000-8000-000000000002",
		EnvironmentID: "30000000-0000-4000-8000-000000000003", Sequence: 1,
		PrimaryBindingID:       "40000000-0000-4000-8000-000000000004",
		TenantSetSHA256:        testDigest,
		OwnerBindingID:         "50000000-0000-4000-8000-000000000005",
		OwnerBindingGeneration: 1,
		DatabaseLineageID:      "60000000-0000-4000-8000-000000000006",
		DatabaseSystemID:       "70000000-0000-4000-8000-000000000007",
		SigningGeneration:      1, RecoveryTrustBundleSHA256: testDigest,
		ComponentInventory: []byte(`{}`),
		DatabaseNow:        time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}}
	_, found, err := testService(fixture).ClaimNext(
		context.Background(), testAttemptID, testLeaseToken,
	)
	if err != nil || !found {
		t.Fatalf("claim durable attempt: found=%v err=%v", found, err)
	}
}
