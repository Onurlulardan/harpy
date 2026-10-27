package providerapp

import (
	"time"

	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
	"github.com/google/uuid"
)

type WebhookAcceptance struct {
	receiptID  uuid.UUID
	provider   domainprovider.Key
	verified   VerifiedWebhook
	body       ProtectedBody
	receivedAt time.Time
	processor  AsyncJob
}

func newWebhookAcceptance(
	receiptID uuid.UUID,
	key domainprovider.Key,
	verified VerifiedWebhook,
	body ProtectedBody,
	receivedAt time.Time,
	processor AsyncJob,
) (WebhookAcceptance, error) {
	canonicalReceived, err := canonicalInstant(receivedAt)
	if err != nil || receiptID == uuid.Nil || !key.Valid() || !verified.valid() ||
		!body.valid() || !processor.valid() ||
		processor.Policy() != WebhookProcessorPolicy() {
		return WebhookAcceptance{}, ErrWebhookDecision
	}
	return WebhookAcceptance{
		receiptID: receiptID, provider: key, verified: verified,
		body: body, receivedAt: canonicalReceived, processor: processor,
	}, nil
}

func (value WebhookAcceptance) ReceiptID() uuid.UUID         { return value.receiptID }
func (value WebhookAcceptance) Provider() domainprovider.Key { return value.provider }
func (value WebhookAcceptance) Verified() VerifiedWebhook    { return value.verified }
func (value WebhookAcceptance) Body() ProtectedBody          { return cloneProtectedBody(value.body) }
func (value WebhookAcceptance) ReceivedAt() time.Time        { return value.receivedAt }
func (value WebhookAcceptance) ProcessorJob() AsyncJob       { return value.processor }

func cloneProtectedBody(value ProtectedBody) ProtectedBody {
	copy, _ := NewProtectedBody(
		value.algorithm, value.nonce, value.ciphertext, value.keyID,
		value.wrappedDEK, value.bodySHA256,
	)
	return copy
}
