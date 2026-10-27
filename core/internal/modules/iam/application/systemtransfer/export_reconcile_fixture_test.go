package systemtransfer

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

type exportReconcileStoreFixture struct {
	observation ExportReconciliationObservation
	claimed     bool
	sealed      int
	missing     int
	retried     int
	parked      int
	record      ExportReconciliationUnknownRecord
	receipt     ExportReconciliationReceipt
}

func (store *exportReconcileStoreFixture) AcquireReconciliation(
	context.Context, ExportReconciliationClaim,
) (ExportReconciliationObservation, bool, error) {
	return store.observation, store.claimed, nil
}

func (store *exportReconcileStoreFixture) PersistReconciliationSealed(
	_ context.Context, _ ExportReconciliationClaim,
	_ ExportReconciliationObservation, _ ExportArtifactReconciliation,
	receipt ExportReconciliationReceipt,
) error {
	store.sealed++
	store.receipt = receipt
	return nil
}

func (store *exportReconcileStoreFixture) PersistReconciliationMissing(
	_ context.Context, _ ExportReconciliationClaim,
	_ ExportReconciliationObservation, _ ExportArtifactReconciliation,
	receipt ExportReconciliationReceipt, _ string,
) error {
	store.missing++
	store.receipt = receipt
	return nil
}

func (store *exportReconcileStoreFixture) ScheduleReconciliationRetry(
	_ context.Context, _ ExportReconciliationClaim,
	_ ExportReconciliationObservation, record ExportReconciliationUnknownRecord,
) error {
	store.retried++
	store.record = record
	return nil
}

func (store *exportReconcileStoreFixture) PersistReconciliationParked(
	_ context.Context, _ ExportReconciliationClaim,
	_ ExportReconciliationObservation, record ExportReconciliationUnknownRecord,
) error {
	store.parked++
	store.record = record
	return nil
}

type exportReconcilerFixture struct {
	result ExportArtifactReconciliation
	err    error
}

func (fixture exportReconcilerFixture) ReconcileExportArtifact(
	context.Context, ExportArtifactEffectStart,
) (ExportArtifactReconciliation, error) {
	return fixture.result, fixture.err
}

type exportBackoffFixture struct{ delay time.Duration }

func (fixture exportBackoffFixture) RetryDelay(int) (time.Duration, error) {
	return fixture.delay, nil
}

func fixtureExportReconciliation(now time.Time) (
	ExportReconciliationClaim, ExportReconciliationObservation,
) {
	attempt := fixtureExportAttemptWithoutTest(now)
	attempt.State = ExportOutcomeUnknown
	attempt.AttemptID = fixtureAttemptID
	attempt.LeaseOwnerID = fixtureLeaseID
	attempt.AttemptGeneration = 1
	effect := ExportArtifactEffectStart{Attempt: attempt,
		EffectID: fixtureAttemptID, EffectReceiptID: fixtureExportReceiptID,
		ObjectKey:     "iam-system-exports/v1/" + attempt.ExportID.String() + "/1.fcart",
		RequestSHA256: sha256.Sum256([]byte("upload-request")), StartedAt: now,
		Draft: ExportArtifactDraftEvidence{KeyID: "iam-export-key-1",
			HeaderSHA256: [32]byte{1}, CiphertextSHA256: [32]byte{2},
			CiphertextBytes: 128, WrappedDEK: make([]byte, 32),
			WrapNonce: [12]byte{1}}}
	claim := ExportReconciliationClaim{TenantID: attempt.TenantID,
		ExportID:     attempt.ExportID,
		JobID:        uuid.MustParse("019be000-0000-7000-8000-000000000020"),
		AttemptID:    uuid.MustParse("019be000-0000-7000-8000-000000000021"),
		LeaseOwnerID: uuid.MustParse("019be000-0000-7000-8000-000000000022"),
		Resource:     attempt.Resource, Attempt: 1,
		MaximumAttempts: ExportReconciliationMaximumAttempts, LeaseGeneration: 1}
	observation := ExportReconciliationObservation{ExportEffect: effect,
		EffectID:        claim.AttemptID,
		EffectReceiptID: uuid.MustParse("019be000-0000-7000-8000-000000000023"),
		RequestSHA256:   exportReconciliationRequestSHA(effect),
		StartedAt:       now.Add(time.Second)}
	return claim, observation
}
