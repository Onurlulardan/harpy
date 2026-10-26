package formrecordarchive

import (
	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func project(value archiveapp.Snapshot, tenantID,
	actorID domain.ID,
) (formrecordarchiveapi.Snapshot, error) {
	state := formrecordarchiveapi.State(value.State)
	if value.TenantID != tenantID || value.ActorID != actorID ||
		!value.ID.UUIDv7() || !value.JobID.UUIDv7() ||
		(value.RetryOfArchiveID.Valid() && !value.RetryOfArchiveID.UUIDv7()) ||
		value.RequestSHA256 == [32]byte{} || !validState(state) {
		return formrecordarchiveapi.Snapshot{}, formrecordarchiveapi.ErrUnavailable
	}
	return formrecordarchiveapi.Snapshot{
		ArchiveID:        formrecordarchiveapi.ID(value.ID),
		JobID:            formrecordarchiveapi.ID(value.JobID),
		RetryOfArchiveID: formrecordarchiveapi.ID(value.RetryOfArchiveID),
		State:            state, RequestSHA256: value.RequestSHA256,
		EntryCount: value.EntryCount, PlaintextBytes: value.PlaintextBytes,
		CiphertextBytes: value.CiphertextBytes, ReadyAt: value.ReadyAt,
		ExpiresAt: value.ExpiresAt,
	}, nil
}

func validState(state formrecordarchiveapi.State) bool {
	switch state {
	case formrecordarchiveapi.StatePending, formrecordarchiveapi.StateCancelled,
		formrecordarchiveapi.StateReady, formrecordarchiveapi.StateExpired,
		formrecordarchiveapi.StatePurged:
		return true
	default:
		return false
	}
}
