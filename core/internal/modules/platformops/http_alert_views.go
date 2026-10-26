package platformops

import (
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type alertChannelView struct {
	ID            string         `json:"id,omitempty"`
	AlertConfigID string         `json:"alert_config_id,omitempty"`
	ChannelType   string         `json:"channel_type"`
	Config        map[string]any `json:"config"`
	Enabled       bool           `json:"enabled"`
}

type alertRuleView struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Enabled         bool               `json:"enabled"`
	ServiceName     *string            `json:"service_name"`
	Condition       string             `json:"condition"`
	CooldownMinutes int                `json:"cooldown_minutes"`
	Channels        []alertChannelView `json:"channels"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type alertEventView struct {
	ID             string     `json:"id"`
	AlertConfigID  *string    `json:"alert_config_id"`
	ServiceName    string     `json:"service_name"`
	PreviousStatus string     `json:"previous_status"`
	CurrentStatus  string     `json:"current_status"`
	ChannelsUsed   []string   `json:"channels_used"`
	Message        string     `json:"message"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

func alertRuleAPI(value platformopsapi.AlertRule) alertRuleView {
	result := alertRuleView{ID: value.ID, Name: value.Name, Description: value.Description,
		Enabled: value.Enabled, Condition: string(value.Condition),
		CooldownMinutes: value.CooldownMinutes, CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt, Channels: make([]alertChannelView, len(value.Channels))}
	if value.ServiceName != "" {
		result.ServiceName = &value.ServiceName
	}
	for index, channel := range value.Channels {
		result.Channels[index] = alertChannelView{ID: channel.ID,
			AlertConfigID: value.ID, ChannelType: channel.Type,
			Config: channel.Config, Enabled: channel.Enabled}
	}
	return result
}

func alertEventAPI(value platformopsapi.AlertEvent) alertEventView {
	result := alertEventView{ID: value.ID, ServiceName: value.ServiceName,
		PreviousStatus: string(value.PreviousStatus), CurrentStatus: string(value.CurrentStatus),
		ChannelsUsed: value.ChannelsUsed, Message: value.Message,
		ResolvedAt: value.ResolvedAt, CreatedAt: value.CreatedAt}
	if value.RuleID != "" {
		result.AlertConfigID = &value.RuleID
	}
	return result
}
