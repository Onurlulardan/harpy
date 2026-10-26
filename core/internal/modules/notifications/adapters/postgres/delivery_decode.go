package postgres

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func scanDelivery(row scanner) (domain.Delivery, error) {
	var value domain.Delivery
	var id, message string
	var provider *string
	err := row.Scan(&id, &message, &provider, &value.Channel, &value.State,
		&value.Attempt, &value.MaximumAttempts, &value.LeaseOwner,
		&value.LeaseGeneration, &value.LeaseUntil, &value.NotBefore,
		&value.OutcomeCode, &value.OutcomeUnknown, &value.Mandatory, &value.Revision)
	if err != nil {
		return domain.Delivery{}, err
	}
	value.ID, err = decodeID(id)
	if err == nil {
		value.MessageID, err = decodeID(message)
	}
	if err == nil && provider != nil {
		value.ProviderConfigID, err = decodeID(*provider)
	}
	if err != nil || !value.Channel.Valid() || value.Channel == domain.ChannelInApp ||
		value.Revision < 1 || value.Attempt < 0 || value.Attempt > value.MaximumAttempts ||
		value.MaximumAttempts != 8 || value.NotBefore.IsZero() {
		return domain.Delivery{}, fmt.Errorf("decode notification delivery")
	}
	return value, nil
}

func nullableID(value domain.ID) any {
	if !value.Valid() {
		return ""
	}
	return value.String()
}
