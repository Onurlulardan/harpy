package deadletter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

func classifyEventReason(
	envelope eventsv1.Envelope,
	consumer eventsv1.ConsumerPolicy,
	evidence DeliveryEvidence,
	consumers eventsv1.ConsumerCatalog,
) (ReasonClass, string, error) {
	if evidence.DeliveryCount < 0 {
		return "", "", fmt.Errorf("broker dead-letter delivery count is invalid")
	}
	switch evidence.Reason {
	case ReasonDeliveryLimit:
		return ReasonDeliveryLimit, EventDeliveryCode, nil
	case ReasonPermanent:
		if _, err := consumers.Resolve(
			consumer.Consumer, consumer.ConsumerVersion,
			envelope.EventType, envelope.EventVersion,
		); err != nil {
			return ReasonUnknownVersion, EventVersionCode, nil
		}
		return ReasonPermanent, EventPermanentCode, nil
	default:
		return "", "", fmt.Errorf("broker dead-letter reason is invalid")
	}
}

func eventRootID(
	eventID string,
	consumer eventsv1.ConsumerPolicy,
	sourceSHA string,
	reason ReasonClass,
) uuid.UUID {
	document := fmt.Sprintf(
		"dead-letter-root.v1\n%s\n%s\n%d\n%s\n%s",
		eventID, consumer.Consumer, consumer.ConsumerVersion, sourceSHA, reason,
	)
	digest := sha256.Sum256([]byte(document))
	bytes := digest[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x80
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return uuid.Must(uuid.FromBytes(bytes))
}

func eventErrorDigest(
	eventID string,
	consumer eventsv1.ConsumerPolicy,
	reasonCode string,
	evidence DeliveryEvidence,
) string {
	document := fmt.Sprintf(
		"broker-dead-letter.v1\n%s\n%s\n%d\n%s\n%d",
		eventID, consumer.Consumer, consumer.ConsumerVersion,
		reasonCode, evidence.DeliveryCount,
	)
	digest := sha256.Sum256([]byte(document))
	return hex.EncodeToString(digest[:])
}
