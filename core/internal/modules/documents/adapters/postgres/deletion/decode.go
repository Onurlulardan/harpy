package deletion

import (
	"time"

	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

func scanTask(value row, tenantID domain.ID,
	notFound rowNotFound,
) (deletionapp.Task, error) {
	var id, targetID, ownerID uuid.UUID
	var reservationID *uuid.UUID
	var targetKind, key, version, action string
	var objectGeneration, leaseGeneration, plainBytes, cipherBytes int64
	var leaseExpires time.Time
	err := value.Scan(&id, &targetKind, &targetID, &key, &version,
		&objectGeneration, &plainBytes, &cipherBytes, &ownerID, &leaseGeneration,
		&leaseExpires, &reservationID, &action)
	if err != nil {
		if notFound(err) {
			return deletionapp.Task{}, deletionapp.ErrConflict
		}
		return deletionapp.Task{}, deletionapp.ErrUnavailable
	}
	if id == uuid.Nil || targetID == uuid.Nil || ownerID == uuid.Nil || key == "" ||
		version == "" || objectGeneration < 1 || leaseGeneration < 1 ||
		plainBytes < 0 || cipherBytes < 1 || leaseExpires.IsZero() ||
		!validAction(deletionapp.BillingAction(action)) {
		return deletionapp.Task{}, deletionapp.ErrUnavailable
	}
	result := deletionapp.Task{ID: domain.ID(id), TenantID: tenantID,
		TargetID: domain.ID(targetID), LeaseOwnerID: domain.ID(ownerID),
		TargetKind: targetKind, ObjectKey: key, ObjectVersion: version,
		ObjectGeneration: uint64(objectGeneration), LeaseGeneration: uint64(leaseGeneration),
		PlaintextBytes: plainBytes, CiphertextBytes: cipherBytes,
		LeaseExpiresAt: leaseExpires.UTC(), BillingAction: deletionapp.BillingAction(action)}
	if reservationID != nil {
		result.BillingReservationID = domain.ID(*reservationID)
	}
	return result, nil
}

func validAction(value deletionapp.BillingAction) bool {
	return value == deletionapp.BillingKeep || value == deletionapp.BillingRelease ||
		value == deletionapp.BillingDecrease
}
