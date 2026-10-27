package application

import (
	"strings"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func validAlertInput(input platformopsapi.AlertRuleInput) bool {
	if !validRecoveryUUID(input.EnvironmentID) || strings.TrimSpace(input.Name) == "" ||
		len(input.Name) > 120 || len(input.Description) > 1000 ||
		len(input.ServiceName) > 100 || input.CooldownMinutes < 1 ||
		input.CooldownMinutes > 10080 || len(input.Channels) > 2 {
		return false
	}
	if input.Condition != platformopsapi.AlertUnhealthy &&
		input.Condition != platformopsapi.AlertDegraded &&
		input.Condition != platformopsapi.AlertAnyChange {
		return false
	}
	seen := make(map[string]struct{}, len(input.Channels))
	for _, channel := range input.Channels {
		if (channel.Type != "email" && channel.Type != "telegram") ||
			len(channel.Config) != 0 {
			return false
		}
		if _, duplicate := seen[channel.Type]; duplicate {
			return false
		}
		seen[channel.Type] = struct{}{}
	}
	return true
}
