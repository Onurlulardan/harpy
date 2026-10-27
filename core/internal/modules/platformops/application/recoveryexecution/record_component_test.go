package recoveryexecution

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestRecordComponentAcceptsCanonicalKind(t *testing.T) {
	fixture := &executionStoreFixture{changed: true}
	err := testService(fixture).RecordComponent(context.Background(),
		ports.RecoveryComponentRecord{
			RecoverySetID: testSetID, AttemptID: testAttemptID,
			LeaseToken: testLeaseToken, Kind: "database_archive",
			ObjectKey: "recovery/database.dump", ObjectVersion: "v1",
			CipherSHA256: testDigest, CiphertextBytes: 1,
		})
	if err != nil {
		t.Fatalf("record component: %v", err)
	}
}
