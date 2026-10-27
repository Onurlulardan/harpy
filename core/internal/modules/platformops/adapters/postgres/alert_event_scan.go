package postgres

import platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"

type alertEventRow interface{ Scan(...any) error }

func scanAlertEvent(row alertEventRow, total *int64) (platformopsapi.AlertEvent, error) {
	var result platformopsapi.AlertEvent
	var ruleID *string
	arguments := []any{&result.ID, &result.EnvironmentID, &ruleID,
		&result.ServiceName, &result.PreviousStatus, &result.CurrentStatus,
		&result.ChannelsUsed, &result.Message, &result.ResolvedAt, &result.CreatedAt}
	if total != nil {
		arguments = append(arguments, total)
	}
	if err := row.Scan(arguments...); err != nil {
		return platformopsapi.AlertEvent{}, err
	}
	if ruleID != nil {
		result.RuleID = *ruleID
	}
	return result, nil
}
