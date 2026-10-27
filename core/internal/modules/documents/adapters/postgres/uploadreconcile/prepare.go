package uploadreconcile

import (
	"context"
	"time"

	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

func (storage *store) Prepare(ctx context.Context, uploadID domain.ID,
	generation uint64, now time.Time,
) (uploadreconcileapp.Plan, error) {
	if !storage.usable(ctx) || !uploadID.UUIDv7() || generation == 0 || now.IsZero() {
		return uploadreconcileapp.Plan{}, uploadreconcileapp.ErrInvalid
	}
	var id, documentID, versionID, reservationID uuid.UUID
	var currentGeneration, plaintextBytes, ciphertextBytes int64
	var uploadState string
	var expiresAt time.Time
	var leaseExpiresAt *time.Time
	var attemptGeneration *int64
	var attemptState, objectKey, objectVersion *string
	var objectGeneration, objectBytes *int64
	var attemptUpdatedAt *time.Time
	err := storage.queryRow(ctx, prepareSQL, uploadID.String(), generation).Scan(
		&id, &documentID, &versionID, &reservationID, &currentGeneration,
		&plaintextBytes, &ciphertextBytes, &uploadState, &expiresAt, &leaseExpiresAt,
		&attemptGeneration, &attemptState, &objectKey, &objectVersion,
		&objectGeneration, &objectBytes, &attemptUpdatedAt)
	if err != nil {
		if storage.notFound(err) {
			return uploadreconcileapp.Plan{}, uploadreconcileapp.ErrConflict
		}
		return uploadreconcileapp.Plan{}, uploadreconcileapp.ErrUnavailable
	}
	state, ok := decodeUploadState(uploadState)
	if !ok || id == uuid.Nil || documentID == uuid.Nil || versionID == uuid.Nil ||
		reservationID == uuid.Nil || currentGeneration < 0 || plaintextBytes < 1 ||
		ciphertextBytes <= plaintextBytes || expiresAt.IsZero() {
		return uploadreconcileapp.Plan{}, uploadreconcileapp.ErrUnavailable
	}
	plan := uploadreconcileapp.Plan{UploadID: domain.ID(id), TenantID: storage.tenantID,
		DocumentID: domain.ID(documentID), VersionID: domain.ID(versionID),
		BillingReservationID: domain.ID(reservationID), Generation: generation,
		PlaintextBytes: plaintextBytes, CiphertextBytes: ciphertextBytes,
		State: state, ExpiresAt: expiresAt.UTC(), ObservedAt: now.UTC()}
	key, keyErr := domain.ContentObjectKey(storage.tenantID, plan.DocumentID, plan.VersionID)
	if keyErr != nil {
		return uploadreconcileapp.Plan{}, uploadreconcileapp.ErrUnavailable
	}
	plan.ObjectKey = key
	if uint64(currentGeneration) != generation ||
		state != domain.UploadReceiving && state != domain.UploadScanning {
		return plan, nil
	}
	if attemptGeneration == nil || *attemptGeneration != int64(generation) ||
		attemptState == nil || attemptUpdatedAt == nil {
		return uploadreconcileapp.Plan{}, uploadreconcileapp.ErrConflict
	}
	stale := state == domain.UploadReceiving && leaseExpiresAt != nil &&
		!leaseExpiresAt.After(now) || state == domain.UploadScanning &&
		!attemptUpdatedAt.Add(time.Hour).After(now)
	if !stale {
		return uploadreconcileapp.Plan{}, uploadreconcileapp.ErrConflict
	}
	switch *attemptState {
	case "RECEIVING":
		plan.Actionable = true
	case "EFFECT_RECORDED":
		if objectKey == nil || objectVersion == nil || objectGeneration == nil ||
			objectBytes == nil || *objectKey != key || *objectGeneration != int64(generation) ||
			*objectBytes != ciphertextBytes || *objectVersion == "" {
			return uploadreconcileapp.Plan{}, uploadreconcileapp.ErrUnavailable
		}
		plan.Actionable, plan.ObjectVersion = true, *objectVersion
	default:
		return plan, nil
	}
	return plan, nil
}

func decodeUploadState(value string) (domain.UploadState, bool) {
	states := map[string]domain.UploadState{
		"CREATED": domain.UploadCreated, "RECEIVING": domain.UploadReceiving,
		"RETRYABLE": domain.UploadRetryable, "SCANNING": domain.UploadScanning,
		"READY": domain.UploadReady, "REJECTED": domain.UploadRejected,
		"CONFLICTED": domain.UploadConflicted, "CANCELLED": domain.UploadCancelled,
		"EXPIRED": domain.UploadExpired,
	}
	state, ok := states[value]
	return state, ok
}
