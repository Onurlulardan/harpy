package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type DeliveryWork struct {
	Delivery domain.Delivery
	Message  domain.Message
	Provider ProviderConfig
	Secret   EncryptedSecret
}

type DeliveryFinalization struct {
	Delivery        domain.Delivery
	Attempt         DeliveryAttempt
	Receipt         *DeliveryReceipt
	ProviderLocator *ProviderLocator
}

type ProviderLocator struct {
	DeliveryID            domain.ID
	ProviderKind          ProviderKind
	ProviderAccountDigest [32]byte
	ExternalMessageDigest [32]byte
	AcceptedAt            time.Time
}

type DeliveryRuntimeStore interface {
	AcquireDelivery(context.Context, domain.ID, domain.ID, string, time.Time) (DeliveryWork, error)
	CommitDelivery(context.Context, domain.ID, DeliveryFinalization, int64) error
}
