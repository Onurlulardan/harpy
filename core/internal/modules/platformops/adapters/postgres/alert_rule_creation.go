package postgres

import (
	"context"
	"fmt"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/google/uuid"
)

func (store *Store) AlertRuleCreation(
	ctx context.Context,
	input platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	channels, err := encodeAlertChannels(input.Channels)
	if err != nil {
		return platformopsapi.AlertRule{}, err
	}
	var result platformopsapi.AlertRule
	err = store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		row := database.QueryRow(ctx, createAlertRuleSQL, uuid.NewString(), input.EnvironmentID,
			input.Name, input.Description, input.Enabled, input.ServiceName,
			input.Condition, input.CooldownMinutes, channels, time.Now().UTC())
		var scanErr error
		result, scanErr = scanAlertRule(row)
		return store.mapDatabaseError(scanErr)
	})
	if err != nil {
		return platformopsapi.AlertRule{}, fmt.Errorf("create alert rule: %w", err)
	}
	return result, nil
}
