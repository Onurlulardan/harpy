package upload

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

type storedSnapshot struct {
	id, actorID, documentID, candidateID, folderID, reservationID uuid.UUID
	expectedCurrent                                               *uuid.UUID
	requestSHA256, contentSHA256, readyResponse, tags             []byte
	expectedBytes, upperBound, generation                         int64
	originalName, declaredMIME, description, state                string
	cleanupRequired                                               bool
	stableError                                                   *string
	readyStatus                                                   *int16
	expiresAt, createdAt, updatedAt                               time.Time
	cancelRequestedAt                                             *time.Time
}

func scanStored(scanner row, value *storedSnapshot) error {
	return scanner.Scan(storedFields(value)...)
}

func storedFields(value *storedSnapshot) []any {
	return []any{&value.id, &value.actorID, &value.documentID, &value.candidateID,
		&value.expectedCurrent, &value.folderID, &value.reservationID, &value.requestSHA256,
		&value.contentSHA256, &value.expectedBytes, &value.upperBound, &value.originalName,
		&value.declaredMIME, &value.description, &value.tags, &value.generation,
		&value.state, &value.cleanupRequired,
		&value.stableError, &value.readyStatus, &value.readyResponse, &value.expiresAt,
		&value.createdAt, &value.updatedAt, &value.cancelRequestedAt}
}

func scanSnapshot(scanner row, tenantID domain.ID,
	notFound rowNotFound,
) (uploadapp.Snapshot, error) {
	stored := storedSnapshot{}
	if err := scanStored(scanner, &stored); err != nil {
		if notFound(err) {
			return uploadapp.Snapshot{}, uploadapp.ErrNotFound
		}
		return uploadapp.Snapshot{}, uploadapp.ErrUnavailable
	}
	return stored.decode(tenantID)
}

func (stored storedSnapshot) decode(tenantID domain.ID) (uploadapp.Snapshot, error) {
	state, ok := decodeState(stored.state)
	var tags []string
	jsonErr := json.Unmarshal(stored.tags, &tags)
	description, normalizedTags, _, metadataOK := foundation.NormalizeItemMetadata(
		stored.description, tags)
	if !ok || stored.id == uuid.Nil || stored.actorID == uuid.Nil ||
		stored.documentID == uuid.Nil || stored.candidateID == uuid.Nil ||
		stored.folderID == uuid.Nil || stored.reservationID == uuid.Nil ||
		len(stored.requestSHA256) != 32 || len(stored.contentSHA256) != 32 ||
		stored.expectedBytes < 1 || stored.upperBound <= stored.expectedBytes ||
		stored.generation < 0 || stored.originalName == "" || stored.declaredMIME == "" ||
		stored.expiresAt.IsZero() || stored.createdAt.IsZero() || stored.updatedAt.IsZero() ||
		jsonErr != nil || !metadataOK || description != stored.description ||
		!slices.Equal(tags, normalizedTags) {
		return uploadapp.Snapshot{}, uploadapp.ErrUnavailable
	}
	result := uploadapp.Snapshot{ID: domain.ID(stored.id), TenantID: tenantID,
		ActorID: domain.ID(stored.actorID), DocumentID: domain.ID(stored.documentID),
		CandidateVersionID: domain.ID(stored.candidateID), FolderID: domain.ID(stored.folderID),
		BillingReservationID: domain.ID(stored.reservationID), ExpectedBytes: stored.expectedBytes,
		CiphertextUpperBound: stored.upperBound, OriginalName: stored.originalName,
		DeclaredMIME: stored.declaredMIME, Description: stored.description,
		Tags: append([]string(nil), tags...), AttemptGeneration: uint64(stored.generation),
		State: state, CleanupRequired: stored.cleanupRequired,
		ExpiresAt: stored.expiresAt.UTC(), CreatedAt: stored.createdAt.UTC(),
		UpdatedAt: stored.updatedAt.UTC(), ReadyResponse: append([]byte(nil), stored.readyResponse...)}
	copy(result.RequestSHA256[:], stored.requestSHA256)
	copy(result.ExpectedContentSHA256[:], stored.contentSHA256)
	if stored.expectedCurrent != nil {
		result.ExpectedCurrentVersionID = domain.ID(*stored.expectedCurrent)
	}
	if stored.stableError != nil {
		result.StableErrorCode = *stored.stableError
	}
	if stored.readyStatus != nil {
		result.ReadyStatus = int(*stored.readyStatus)
	}
	if stored.cancelRequestedAt != nil {
		result.CancelRequestedAt = stored.cancelRequestedAt.UTC()
	}
	return result, nil
}

func decodeState(value string) (domain.UploadState, bool) {
	states := map[string]domain.UploadState{"CREATED": domain.UploadCreated,
		"RECEIVING": domain.UploadReceiving, "RETRYABLE": domain.UploadRetryable,
		"SCANNING": domain.UploadScanning, "READY": domain.UploadReady,
		"REJECTED": domain.UploadRejected, "CONFLICTED": domain.UploadConflicted,
		"CANCELLED": domain.UploadCancelled, "EXPIRED": domain.UploadExpired}
	state, ok := states[value]
	return state, ok
}
