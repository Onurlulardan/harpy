package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type ProviderReceiptOutcome string

const (
	ReceiptDelivered  ProviderReceiptOutcome = "delivered"
	ReceiptBounced    ProviderReceiptOutcome = "bounced"
	ReceiptComplained ProviderReceiptOutcome = "complained"
)

type AuthenticatedProviderReceipt struct {
	ID, DeliveryID          domain.ID
	ProviderReferenceDigest [32]byte
	TargetDigest            [32]byte
	Outcome                 ProviderReceiptOutcome
	SignatureKeyID          string
	ReceivedAt              time.Time
}

type ProviderReceiptRecorder interface {
	PersistProviderReceipt(context.Context, AuthenticatedProviderReceipt) (bool, error)
}
