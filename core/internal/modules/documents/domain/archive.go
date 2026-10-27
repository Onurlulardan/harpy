package domain

import (
	"errors"
	"time"
)

var (
	ErrArchiveInvalid    = errors.New("documents archive input is invalid")
	ErrArchiveLimit      = errors.New("documents archive limit exceeded")
	ErrArchiveTransition = errors.New("documents archive transition is invalid")
)

const (
	MaximumArchiveFiles             = 10_000
	MaximumArchivePlaintext         = int64(10 << 30)
	MaximumArchiveArtifactPlaintext = MaximumArchivePlaintext + int64(16<<20)
	ArchiveKeep                     = 24 * time.Hour
)

type ArchiveState string

const (
	ArchivePending   ArchiveState = "pending"
	ArchiveCancelled ArchiveState = "cancelled"
	ArchiveReady     ArchiveState = "ready"
	ArchiveExpired   ArchiveState = "expired"
	ArchivePurged    ArchiveState = "purged"
)

type ArchiveEntry struct {
	DocumentID, VersionID ID
	Order                 uint32
	PlaintextBytes        int64
	PlaintextSHA256       [32]byte
}

type Archive struct {
	ID, TenantID, ActorID, JobID, RetryOfArchiveID ID
	RequestSHA256, EntriesSHA256                   [32]byte
	AuthorizationRevision                          uint64
	State                                          ArchiveState
	EntryCount                                     int
	PlaintextBytes, CiphertextBytes                int64
	ArtifactSHA256, DeleteReceiptSHA256            [32]byte
	ReadyAt, ExpiresAt, PurgedAt                   time.Time
}

func NewArchive(id, tenantID, actorID, jobID, retryOf ID,
	requestSHA256 [32]byte, authorizationRevision uint64,
) (Archive, error) {
	if !id.UUIDv7() || !tenantID.Valid() || !actorID.Valid() || !jobID.UUIDv7() ||
		retryOf.Valid() && !retryOf.UUIDv7() || requestSHA256 == [32]byte{} ||
		authorizationRevision == 0 {
		return Archive{}, ErrArchiveInvalid
	}
	return Archive{ID: id, TenantID: tenantID, ActorID: actorID, JobID: jobID,
		RetryOfArchiveID: retryOf, RequestSHA256: requestSHA256,
		AuthorizationRevision: authorizationRevision, State: ArchivePending}, nil
}

func (archive Archive) Pin(entries []ArchiveEntry, digest [32]byte) (Archive, error) {
	if archive.State != ArchivePending || archive.EntryCount != 0 ||
		len(entries) == 0 || digest == [32]byte{} {
		return Archive{}, ErrArchiveTransition
	}
	if len(entries) > MaximumArchiveFiles {
		return Archive{}, ErrArchiveLimit
	}
	seen := make(map[ID]struct{}, len(entries))
	var total int64
	for index, entry := range entries {
		if !entry.DocumentID.UUIDv7() || !entry.VersionID.UUIDv7() ||
			entry.Order != uint32(index) || entry.PlaintextBytes < 1 ||
			entry.PlaintextSHA256 == [32]byte{} {
			return Archive{}, ErrArchiveInvalid
		}
		if _, exists := seen[entry.VersionID]; exists {
			return Archive{}, ErrArchiveInvalid
		}
		seen[entry.VersionID] = struct{}{}
		if entry.PlaintextBytes > MaximumArchivePlaintext-total {
			return Archive{}, ErrArchiveLimit
		}
		total += entry.PlaintextBytes
	}
	archive.EntryCount, archive.PlaintextBytes = len(entries), total
	archive.EntriesSHA256 = digest
	return archive, nil
}

func (archive Archive) MarkReady(now time.Time, ciphertextBytes int64,
	artifactSHA256 [32]byte,
) (Archive, error) {
	if archive.State != ArchivePending || archive.EntryCount == 0 || !validInstant(now) ||
		ciphertextBytes < 1 || artifactSHA256 == [32]byte{} {
		return Archive{}, ErrArchiveTransition
	}
	archive.State, archive.ReadyAt = ArchiveReady, now.UTC()
	archive.ExpiresAt, archive.CiphertextBytes = archive.ReadyAt.Add(ArchiveKeep), ciphertextBytes
	archive.ArtifactSHA256 = artifactSHA256
	return archive, nil
}

func (archive Archive) Expire(now time.Time) (Archive, error) {
	if archive.State != ArchiveReady || !validInstant(now) || now.Before(archive.ExpiresAt) {
		return Archive{}, ErrArchiveTransition
	}
	archive.State = ArchiveExpired
	return archive, nil
}

func (archive Archive) ConfirmDelete(now time.Time, receipt [32]byte) (Archive, error) {
	if archive.State != ArchiveExpired || !validInstant(now) || receipt == [32]byte{} {
		return Archive{}, ErrArchiveTransition
	}
	archive.State, archive.PurgedAt = ArchivePurged, now.UTC()
	archive.DeleteReceiptSHA256 = receipt
	return archive, nil
}
