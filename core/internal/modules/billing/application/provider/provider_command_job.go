package providerapp

import (
	"time"

	"github.com/google/uuid"
)

func NewProviderCommandJob(
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
	identity := provider + ":" + commandID.String()
	return newAsyncJob(
		ProviderCommandPolicy(),
		"billing-provider-command/"+provider+"/"+commandID.String(),
		payloadSHA256, runAt,
		"billing:provider:command:"+identity,
		"billing:provider:command:"+identity,
	)
}
