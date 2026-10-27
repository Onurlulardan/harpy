package archive

import (
	"strings"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) scanSnapshot(value row) (archiveapp.Snapshot, error) {
	var result archiveapp.Snapshot
	var id, actor, job, reservation string
	var retry, objectKey, objectVersion, keyID dbText
	var state string
	var request, entries, artifact, ciphertext, wrapped []byte
	var artifactBytes, cipherBytes, generation dbInt64
	var readyAt, expiresAt dbInstant
	err := value.Scan(&id, &actor, &result.AuthorizationRevision, &job, &retry,
		&request, &reservation, &state, &result.EntryCount, &result.PlaintextBytes,
		&artifactBytes, &cipherBytes, &entries, &artifact, &ciphertext,
		&objectKey, &objectVersion, &generation, &keyID, &wrapped,
		&readyAt, &expiresAt, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		if storage.notFound(err) {
			return archiveapp.Snapshot{}, archiveapp.ErrNotFound
		}
		return archiveapp.Snapshot{}, archiveapp.ErrUnavailable
	}
	var parseErr error
	if result.ID, parseErr = domain.ParseID(id); parseErr == nil {
		result.ActorID, parseErr = domain.ParseID(actor)
	}
	if parseErr == nil {
		result.JobID, parseErr = domain.ParseID(job)
	}
	if parseErr == nil {
		result.BillingReservationID, parseErr = domain.ParseID(reservation)
	}
	if parseErr != nil || len(request) != 32 || !decodeArchiveState(state, &result.State) {
		return archiveapp.Snapshot{}, archiveapp.ErrUnavailable
	}
	result.TenantID = storage.tenantID
	copy(result.RequestSHA256[:], request)
	copy(result.EntriesSHA256[:], entries)
	copy(result.ArtifactSHA256[:], artifact)
	copy(result.CiphertextSHA256[:], ciphertext)
	result.ArtifactPlaintextBytes, result.CiphertextBytes = artifactBytes.Int64, cipherBytes.Int64
	result.ObjectKey, result.ObjectVersion, result.KeyID = objectKey.Text,
		objectVersion.Text, keyID.Text
	result.ObjectGeneration = uint64(generation.Int64)
	result.WrappedDEK = append([]byte(nil), wrapped...)
	result.ReadyAt, result.ExpiresAt = readyAt.Time, expiresAt.Time
	if retry.Valid {
		result.RetryOfArchiveID, parseErr = domain.ParseID(retry.Text)
	}
	if parseErr != nil {
		return archiveapp.Snapshot{}, archiveapp.ErrUnavailable
	}
	return result, nil
}

func decodeArchiveState(value string, target *domain.ArchiveState) bool {
	switch strings.ToUpper(value) {
	case "PENDING":
		*target = domain.ArchivePending
	case "CANCELLED":
		*target = domain.ArchiveCancelled
	case "READY":
		*target = domain.ArchiveReady
	case "EXPIRED":
		*target = domain.ArchiveExpired
	case "PURGED":
		*target = domain.ArchivePurged
	default:
		return false
	}
	return true
}
