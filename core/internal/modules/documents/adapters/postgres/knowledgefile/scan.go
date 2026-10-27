package knowledgefile

import (
	"time"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

func scanLease(source row, tenantID domain.ID) (knowledgeapp.Lease, error) {
	return scanLeaseRow(source, tenantID, false)
}

func scanExpiredLease(source row, tenantID domain.ID) (knowledgeapp.Lease, error) {
	value, err := scanLeaseRow(source, tenantID, true)
	if err != nil || value.ItemRowVersion == 0 {
		return knowledgeapp.Lease{}, knowledgeapi.ErrUnavailable
	}
	return value, nil
}

func scanLeaseRow(source row, tenantID domain.ID,
	withRowVersion bool,
) (knowledgeapp.Lease, error) {
	var leaseID, uploadID, actorID, itemID, versionID uuid.UUID
	var referenceID, sourceID *uuid.UUID
	var expectedBytes int64
	var digest []byte
	var filename, declaredMIME, state string
	var detectedMIME, revision *string
	var scannedAt, adoptedAt, releasedAt *time.Time
	var expiresAt time.Time
	values := []any{&leaseID, &uploadID, &actorID, &itemID, &versionID,
		&referenceID, &sourceID, &expectedBytes, &digest, &filename, &declaredMIME,
		&detectedMIME, &revision, &scannedAt, &state, &expiresAt, &adoptedAt, &releasedAt}
	var rowVersion int64
	if withRowVersion {
		values = append(values, &rowVersion)
	}
	err := source.Scan(values...)
	if err != nil {
		return knowledgeapp.Lease{}, err
	}
	value := knowledgeapp.Lease{LeaseID: domain.ID(leaseID), UploadID: domain.ID(uploadID),
		TenantID: tenantID, ActorID: domain.ID(actorID), ItemID: domain.ID(itemID),
		VersionID: domain.ID(versionID), ExpectedBytes: expectedBytes,
		Filename: filename, MIME: declaredMIME, State: state,
		ItemRowVersion: uint64(rowVersion), ExpiresAt: expiresAt.UTC()}
	if len(digest) == 32 {
		copy(value.ExpectedSHA256[:], digest)
	}
	if referenceID != nil {
		value.ReferenceID = domain.ID(*referenceID)
	}
	if sourceID != nil {
		value.SourceID = domain.ID(*sourceID)
	}
	if detectedMIME != nil {
		value.DetectedMIME = *detectedMIME
	}
	if revision != nil {
		value.ScanRevision = *revision
	}
	if scannedAt != nil {
		value.ScannedAt = scannedAt.UTC()
	}
	if rowVersion < 0 || !validLease(value) {
		return knowledgeapp.Lease{}, knowledgeapi.ErrUnavailable
	}
	return value, nil
}

func validLease(value knowledgeapp.Lease) bool {
	base := value.LeaseID.UUIDv7() && value.UploadID == value.LeaseID &&
		value.TenantID.Valid() && value.ActorID.Valid() && value.ItemID.UUIDv7() &&
		value.VersionID.UUIDv7() && value.ExpectedBytes > 0 &&
		value.ExpectedSHA256 != [32]byte{} && value.Filename != "" && value.MIME != "" &&
		!value.ExpiresAt.IsZero()
	if !base {
		return false
	}
	switch value.State {
	case "PENDING", "EXPIRED":
		return !value.ReferenceID.Valid() && !value.SourceID.Valid()
	case "STAGED":
		return value.DetectedMIME != "" && value.ScanRevision != "" &&
			!value.ScannedAt.IsZero() && !value.ReferenceID.Valid()
	case "ADOPTED", "RELEASED":
		return value.DetectedMIME != "" && value.ScanRevision != "" &&
			!value.ScannedAt.IsZero() && value.ReferenceID.UUIDv7() && value.SourceID.UUIDv7()
	default:
		return false
	}
}
