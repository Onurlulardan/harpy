package uploadreconcile

import (
	"context"
	"time"

	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

func (storage *store) PersistFinalization(ctx context.Context, plan uploadreconcileapp.Plan,
	version string, deletionID domain.ID, now time.Time,
) (uploadreconcileapp.Finalization, error) {
	found := version != ""
	if !storage.usable(ctx) || plan.TenantID != storage.tenantID ||
		!plan.UploadID.UUIDv7() || plan.Generation == 0 || plan.ObjectKey == "" ||
		plan.PlaintextBytes < 1 || plan.CiphertextBytes <= plan.PlaintextBytes ||
		found != deletionID.UUIDv7() || now.IsZero() {
		return uploadreconcileapp.Finalization{}, uploadreconcileapp.ErrInvalid
	}
	var uploadID, reservationID uuid.UUID
	var state string
	var release bool
	var versionValue, deletionValue any
	if found {
		versionValue, deletionValue = version, deletionID.String()
	}
	err := storage.queryRow(ctx, finalizeSQL, plan.UploadID.String(), plan.Generation,
		plan.ObjectKey, versionValue, plan.PlaintextBytes, plan.CiphertextBytes,
		deletionValue, now.UTC()).Scan(&uploadID, &reservationID, &state, &release)
	if err != nil {
		if storage.notFound(err) {
			return uploadreconcileapp.Finalization{}, uploadreconcileapp.ErrConflict
		}
		return uploadreconcileapp.Finalization{}, uploadreconcileapp.ErrUnavailable
	}
	decoded, ok := decodeUploadState(state)
	if !ok || uploadID == uuid.Nil || reservationID == uuid.Nil {
		return uploadreconcileapp.Finalization{}, uploadreconcileapp.ErrUnavailable
	}
	return uploadreconcileapp.Finalization{UploadID: domain.ID(uploadID),
		BillingReservationID: domain.ID(reservationID), State: decoded,
		ReleaseReservation: release}, nil
}
