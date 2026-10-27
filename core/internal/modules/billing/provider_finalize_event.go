package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

const providerPaymentEventSchemaSHA256 = "3ce299b97ab5b34969096bd0ebc9e2631bb4ed10ce88bffd56fdc5336edcbb00"

func providerPaymentEvent(
	causationID uuid.UUID,
	result providerapp.Finalization,
	nextID providerapp.CommandIDSource,
) (platformasync.DatabaseJobEventInput, error) {
	eventID, err := nextID()
	if err != nil || eventID == uuid.Nil ||
		eventID.Version() != uuid.Version(7) {
		return platformasync.DatabaseJobEventInput{}, providerapp.ErrConfiguration
	}
	payload := []byte(fmt.Sprintf(
		"{\"provider_command_id\":\"%s\",\"provider_command_revision\":%d}",
		result.CommandID, result.Revision,
	))
	digest := sha256.Sum256(payload)
	effectKey := fmt.Sprintf(
		"billing.payment.changed:%s:%d", result.CommandID, result.Revision,
	)
	return platformasync.DatabaseJobEventInput{
		EventID: eventID, EventType: "billing.payment.changed", EventVersion: 1,
		AggregateType: "billing.payment", AggregateID: result.CommandID,
		AggregateVersion: result.Revision, EventSequence: result.Revision,
		OccurredAt: result.OccurredAt, ActorID: result.ActorID,
		CorrelationID: result.CommandID.String(), CausationID: causationID,
		IdempotencyKey:               eventID.String(),
		BusinessEffectIdempotencyKey: effectKey,
		PayloadSchemaSHA256:          providerPaymentEventSchemaSHA256,
		PayloadSHA256:                hex.EncodeToString(digest[:]), Payload: payload,
	}, nil
}
