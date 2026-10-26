package platformops

import (
	"net/http"
	"strings"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

type alertChannelDocument struct {
	ChannelType string         `json:"channel_type"`
	Config      map[string]any `json:"config"`
	Enabled     bool           `json:"enabled"`
}

type alertRuleDocument struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Enabled         bool                   `json:"enabled"`
	ServiceName     *string                `json:"service_name"`
	Condition       string                 `json:"condition"`
	CooldownMinutes int                    `json:"cooldown_minutes"`
	Channels        []alertChannelDocument `json:"channels"`
}

func decodeAlertRule(
	w http.ResponseWriter,
	r *http.Request,
	environmentID string,
) (platformopsapi.AlertRuleInput, error) {
	var document alertRuleDocument
	if err := httpinput.DecodeJSON(w, r, 32768, &document); err != nil {
		return platformopsapi.AlertRuleInput{}, err
	}
	input := platformopsapi.AlertRuleInput{EnvironmentID: environmentID,
		Name: strings.TrimSpace(document.Name), Description: document.Description,
		Enabled: document.Enabled, Condition: platformopsapi.AlertCondition(document.Condition),
		CooldownMinutes: document.CooldownMinutes,
		Channels:        make([]platformopsapi.AlertChannel, len(document.Channels))}
	if document.ServiceName != nil {
		input.ServiceName = strings.TrimSpace(*document.ServiceName)
	}
	for index, channel := range document.Channels {
		input.Channels[index] = platformopsapi.AlertChannel{Type: channel.ChannelType,
			Config: channel.Config, Enabled: channel.Enabled}
	}
	return input, nil
}
