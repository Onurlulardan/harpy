package providerapp

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
	"github.com/google/uuid"
)

func newWebhookProcessorJob(
	receiptID uuid.UUID,
	bodySHA256 [32]byte,
	runAt time.Time,
) (AsyncJob, error) {
	if receiptID == uuid.Nil || bodySHA256 == [32]byte{} {
		return AsyncJob{}, ErrWebhookDecision
	}
	return newAsyncJob(
		WebhookProcessorPolicy(),
		"billing-webhook-receipt/"+receiptID.String(),
		hex.EncodeToString(bodySHA256[:]), runAt,
		"billing:webhook:process:"+receiptID.String(), "",
	)
}

func newProviderReconciliationJob(
	receiptID uuid.UUID,
	key domainprovider.Key,
	verified VerifiedWebhook,
	eventSHA256 [32]byte,
	runAt time.Time,
) (AsyncJob, error) {
	if receiptID == uuid.Nil || !key.Valid() || !verified.valid() ||
		eventSHA256 == [32]byte{} {
		return AsyncJob{}, ErrWebhookDecision
	}
	identity := fmt.Sprintf(
		"%s\x00%s\x00%s\x00%s", key, verified.AccountID(),
		verified.ObjectKind(), verified.ObjectID(),
	)
	digest := sha256.Sum256([]byte(identity))
	stableKey := hex.EncodeToString(digest[:])
	return newAsyncJob(
		ProviderReconciliationPolicy(),
		"billing-provider-reconciliation/"+string(key)+"/"+receiptID.String(),
		hex.EncodeToString(eventSHA256[:]), runAt,
		"billing:provider:reconcile:"+stableKey,
		"billing:provider:reconcile:"+stableKey,
	)
}
