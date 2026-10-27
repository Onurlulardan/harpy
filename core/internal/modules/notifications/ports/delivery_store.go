package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type DeliveryStore interface {
	InsertDelivery(context.Context, domain.Delivery) error
	InsertAttempt(context.Context, DeliveryAttempt) error
	InsertReceipt(context.Context, DeliveryReceipt) error
}

type DeliveryAttempt struct {
	ID, DeliveryID              domain.ID
	Number                      int
	LeaseGeneration             int64
	ProviderKind, ResponseClass string
	ProviderReferenceDigest     [32]byte
	StatusCode, ErrorCode       string
	StartedAt, FinishedAt       time.Time
}

type DeliveryReceipt struct {
	ID, DeliveryID          domain.ID
	ProviderReferenceDigest [32]byte
	Outcome, SignatureKeyID string
	ReceivedAt              time.Time
}
