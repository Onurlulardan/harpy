package deletion

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Task struct {
	ID, TenantID, TargetID, LeaseOwnerID domain.ID
	BillingReservationID                 domain.ID
	TargetKind, ObjectKey, ObjectVersion string
	ObjectGeneration, LeaseGeneration    uint64
	PlaintextBytes, CiphertextBytes      int64
	LeaseExpiresAt                       time.Time
	BillingAction                        BillingAction
}

type BillingAction string

const (
	BillingKeep     BillingAction = "keep"
	BillingRelease  BillingAction = "release"
	BillingDecrease BillingAction = "decrease"
)

type DeleteReceipt struct {
	ProviderRequestID           string
	DeleteSHA256, AbsenceSHA256 [32]byte
	VerifiedAt                  time.Time
}

type Result struct {
	Task
	DeleteReceipt
	BillingReceiptID domain.ID
	ConfirmedAt      time.Time
}
