package providerapp

import (
	"time"

	"github.com/google/uuid"
)

func NewProviderFinalizeJob(
	commandID uuid.UUID,
	payloadSHA256 string,
	runAt time.Time,
) (AsyncJob, error) {
	if commandID == uuid.Nil || commandID.Version() != uuid.Version(7) {
		return AsyncJob{}, ErrWebhookDecision
	}
	identity := commandID.String()
	return newAsyncJob(
		ProviderFinalizePolicy(),
		"billing-provider-finalize/"+identity,
		payloadSHA256,
		runAt,
		"billing:provider:finalize:"+identity,
		"",
	)
}
