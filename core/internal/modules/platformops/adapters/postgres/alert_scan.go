package postgres

import (
	"encoding/json"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type alertRow interface{ Scan(...any) error }

type alertChannelDocument struct {
	ID          string         `json:"id,omitempty"`
	ChannelType string         `json:"channel_type"`
	Config      map[string]any `json:"config"`
	Enabled     bool           `json:"enabled"`
}

func encodeAlertChannels(channels []platformopsapi.AlertChannel) ([]byte, error) {
	document := make([]alertChannelDocument, len(channels))
	for index, channel := range channels {
		document[index] = alertChannelDocument{ID: channel.ID, ChannelType: channel.Type,
			Config: channel.Config, Enabled: channel.Enabled}
	}
	return json.Marshal(document)
}

func scanAlertRule(row alertRow) (platformopsapi.AlertRule, error) {
	var result platformopsapi.AlertRule
	var serviceName *string
	var channels []byte
	if err := row.Scan(&result.ID, &result.EnvironmentID, &result.Name,
		&result.Description, &result.Enabled, &serviceName, &result.Condition,
		&result.CooldownMinutes, &channels, &result.CreatedAt, &result.UpdatedAt); err != nil {
		return platformopsapi.AlertRule{}, err
	}
	if serviceName != nil {
		result.ServiceName = *serviceName
	}
	var document []alertChannelDocument
	if err := json.Unmarshal(channels, &document); err != nil {
		return platformopsapi.AlertRule{}, fmt.Errorf("decode alert channels: %w", err)
	}
	result.Channels = make([]platformopsapi.AlertChannel, len(document))
	for index, channel := range document {
		result.Channels[index] = platformopsapi.AlertChannel{ID: channel.ID,
			Type: channel.ChannelType, Config: channel.Config, Enabled: channel.Enabled}
	}
	return result, nil
}
