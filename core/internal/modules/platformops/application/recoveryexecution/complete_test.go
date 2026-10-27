package recoveryexecution

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestCompleteRequiresSuccessfulExactTransition(t *testing.T) {
	fixture := &executionStoreFixture{}
	record := ports.RecoveryCompletionRecord{
		RecoverySetID: testSetID, AttemptID: testAttemptID, LeaseToken: testLeaseToken,
		ManifestSHA256: testDigest, ManifestBytes: []byte("{}"),
		DatabaseSystemID: "70000000-0000-4000-8000-000000000007",
	}
	if !errors.Is(testService(fixture).Complete(context.Background(), record), ErrConflict) {
		t.Fatal("unchanged completion must conflict")
	}
	fixture.changed = true
	if err := testService(fixture).Complete(context.Background(), record); err != nil {
		t.Fatalf("complete: %v", err)
	}
}
