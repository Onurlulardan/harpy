package s3object

import (
	"testing"
	"time"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func TestReconciliationEvidenceBindsEveryObservationAndExpectedDigest(t *testing.T) {
	modified := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	expected := artifactobject.ReconcileExpected{Key: testObjectKey,
		CiphertextBytes: 7, CiphertextSHA256: [32]byte{1}, HeaderSHA256: [32]byte{2}}
	before := reconciliationSnapshot{present: true, version: "version-1",
		bytes: 7, lastModified: modified, requestID: "list-before"}
	after := before
	after.requestID = "list-after"
	baseline := reconcileEvidenceSHA(expected, before, after,
		artifactobject.ReconciliationSealed, "head-request")
	changedObservation := after
	changedObservation.requestID = "other-list"
	changedExpected := expected
	changedExpected.HeaderSHA256 = [32]byte{3}
	if baseline == reconcileEvidenceSHA(expected, before, changedObservation,
		artifactobject.ReconciliationSealed, "head-request") ||
		baseline == reconcileEvidenceSHA(changedExpected, before, after,
			artifactobject.ReconciliationSealed, "head-request") {
		t.Fatal("reconciliation evidence did not bind all observations")
	}
}
