package uploadreconcile

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const StateRetryable = domain.UploadRetryable

type Plan struct {
	UploadID, TenantID, DocumentID, VersionID domain.ID
	BillingReservationID                      domain.ID
	Generation                                uint64
	ObjectKey, ObjectVersion                  string
	PlaintextBytes, CiphertextBytes           int64
	State                                     domain.UploadState
	Actionable                                bool
	ExpiresAt, ObservedAt                     time.Time
}

type Finalization struct {
	UploadID, BillingReservationID domain.ID
	State                          domain.UploadState
	ReleaseReservation             bool
}

type Result struct {
	UploadID, DeletionID domain.ID
	State                domain.UploadState
	ObjectFound, Noop    bool
}
