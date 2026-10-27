package providerapp

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

func NewProviderCommandReconciliationJob(
	provider string,
	commandID uuid.UUID,
	payloadSHA256 string,
	runAt time.Time,
) (AsyncJob, error) {
	name, err := CanonicalProviderName(provider)
	if err != nil || name != provider || commandID == uuid.Nil ||
		commandID.Version() != uuid.Version(7) {
		return AsyncJob{}, ErrWebhookDecision
	}
	identity := sha256.Sum256([]byte(provider + "\x00" + commandID.String()))
	stableKey := hex.EncodeToString(identity[:])
	return newAsyncJob(
		ProviderReconciliationPolicy(),
		"billing-provider-reconciliation/"+provider+"/"+commandID.String(),
		payloadSHA256, runAt,
		"billing:provider:reconcile:"+stableKey,
		"billing:provider:reconcile:"+stableKey,
	)
}
