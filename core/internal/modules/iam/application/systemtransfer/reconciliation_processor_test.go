package systemtransfer

import (
	"errors"
	"testing"
	"time"
)

func TestExportReconciliationProcessorPersistsSealedReceipt(t *testing.T) {
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	claim, observation := fixtureExportReconciliation(now)
	artifact := ExportArtifactMetadata{ObjectKey: observation.ExportEffect.ObjectKey,
		ObjectVersion: "version-1", KeyID: observation.ExportEffect.Draft.KeyID,
		ProviderRequestID: "upload-request-1", EffectID: observation.ExportEffect.EffectID,
		EffectReceiptID:  observation.ExportEffect.EffectReceiptID,
		HeaderSHA256:     observation.ExportEffect.Draft.HeaderSHA256,
		CiphertextSHA256: observation.ExportEffect.Draft.CiphertextSHA256,
		RequestSHA256:    observation.ExportEffect.RequestSHA256,
		ResultSHA256:     [32]byte{3}, CiphertextBytes: 128,
		WrappedDEK: make([]byte, 32), WrapNonce: [12]byte{1},
		AttemptGeneration: 1, SealedAt: now.Add(time.Second)}
	result := ExportArtifactReconciliation{State: ExportArtifactSealed,
		Artifact: artifact, ProviderRequestID: "reconcile-head-1",
		EvidenceSHA256: [32]byte{4}, CheckedAt: now.Add(2 * time.Second)}
	store := &exportReconcileStoreFixture{observation: observation, claimed: true}
	processor, err := NewSystemExportReconciliationProcessor(store,
		exportReconcilerFixture{result: result}, exportClockFixture{now: now.Add(3 * time.Second)},
		exportBackoffFixture{delay: time.Second})
	if err != nil || processor.ProcessExportReconciliation(t.Context(), claim) != nil {
		t.Fatalf("processor error = %v", err)
	}
	if store.sealed != 1 || store.missing+store.retried+store.parked != 0 ||
		store.receipt.ResultSHA256 != result.EvidenceSHA256 ||
		store.receipt.ObservedAt != result.CheckedAt {
		t.Fatalf("store = %#v", store)
	}
}

func TestExportReconciliationProcessorTerminatesVerifiedMissing(t *testing.T) {
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	claim, observation := fixtureExportReconciliation(now)
	result := ExportArtifactReconciliation{State: ExportArtifactMissing,
		ProviderRequestID: "reconcile-head-404", EvidenceSHA256: [32]byte{5},
		CheckedAt: now.Add(2 * time.Second)}
	store := &exportReconcileStoreFixture{observation: observation, claimed: true}
	processor, _ := NewSystemExportReconciliationProcessor(store,
		exportReconcilerFixture{result: result}, exportClockFixture{now: now.Add(3 * time.Second)},
		exportBackoffFixture{delay: time.Second})
	if err := processor.ProcessExportReconciliation(t.Context(), claim); err != nil ||
		store.missing != 1 || store.sealed+store.retried+store.parked != 0 {
		t.Fatalf("error = %v, store = %#v", err, store)
	}
}

func TestExportReconciliationProcessorBoundsUnknownThenParks(t *testing.T) {
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name    string
		attempt int
		retry   int
		park    int
		code    string
	}{{"retry", 2, 1, 0, ExportReconciliationUnknownCode},
		{"park", 8, 0, 1, ExportReconciliationParkedCode}} {
		t.Run(test.name, func(t *testing.T) {
			claim, observation := fixtureExportReconciliation(now)
			claim.Attempt, claim.LeaseGeneration = test.attempt, int64(test.attempt)
			store := &exportReconcileStoreFixture{observation: observation, claimed: true}
			processor, _ := NewSystemExportReconciliationProcessor(store,
				exportReconcilerFixture{err: errors.New("provider unavailable")},
				exportClockFixture{now: now.Add(3 * time.Second)},
				exportBackoffFixture{delay: 4 * time.Second})
			if err := processor.ProcessExportReconciliation(t.Context(), claim); err != nil ||
				store.retried != test.retry || store.parked != test.park ||
				store.record.Code != test.code {
				t.Fatalf("error = %v, store = %#v", err, store)
			}
		})
	}
}

func TestExportReconciliationRetryRecordEnforcesAttemptSpecificJitterBound(t *testing.T) {
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	claim, observation := fixtureExportReconciliation(now)
	record := ExportReconciliationUnknownRecord{
		Code: ExportReconciliationUnknownCode,
		FailureSHA256: exportReconciliationFailureSHA(
			claim, observation, ExportReconciliationUnknownCode),
		ObservedAt: now.Add(2 * time.Second), RetryAt: now.Add(8 * time.Second),
	}
	if record.Validate(claim, observation) == nil {
		t.Fatal("attempt-one retry exceeded the five-second full-jitter bound")
	}
}
