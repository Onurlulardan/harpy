package providerapp

import (
	"time"

	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
	"github.com/google/uuid"
)

type StoredWebhook struct {
	receiptID  uuid.UUID
	tenantID   uuid.UUID
	provider   domainprovider.Key
	verified   VerifiedWebhook
	body       ProtectedBody
	receivedAt time.Time
}

func NewStoredWebhook(
	receiptID uuid.UUID,
	tenantID uuid.UUID,
	key domainprovider.Key,
	verified VerifiedWebhook,
	body ProtectedBody,
	receivedAt time.Time,
) (StoredWebhook, error) {
	canonicalReceived, err := canonicalInstant(receivedAt)
	if err != nil || receiptID == uuid.Nil || tenantID == uuid.Nil ||
		!key.Valid() || !verified.valid() || !body.valid() {
		return StoredWebhook{}, ErrWebhookProcessing
	}
	return StoredWebhook{
		receiptID: receiptID, tenantID: tenantID, provider: key,
		verified: verified, body: cloneProtectedBody(body), receivedAt: canonicalReceived,
	}, nil
}

func (value StoredWebhook) ReceiptID() uuid.UUID         { return value.receiptID }
func (value StoredWebhook) TenantID() uuid.UUID          { return value.tenantID }
func (value StoredWebhook) Provider() domainprovider.Key { return value.provider }
func (value StoredWebhook) Verified() VerifiedWebhook    { return value.verified }
func (value StoredWebhook) Body() ProtectedBody          { return cloneProtectedBody(value.body) }
func (value StoredWebhook) ReceivedAt() time.Time        { return value.receivedAt }

func (value StoredWebhook) valid() bool {
	validated, err := NewStoredWebhook(
		value.receiptID, value.tenantID, value.provider, value.verified,
		value.body, value.receivedAt,
	)
	return err == nil && validated.receiptID == value.receiptID &&
		validated.tenantID == value.tenantID && validated.provider == value.provider &&
		protectedBodiesEqual(validated.body, value.body)
}
