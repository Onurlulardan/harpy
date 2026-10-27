package application

import (
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestAlertInputAcceptsOnlySecretFreeUniqueChannelTypes(t *testing.T) {
	t.Parallel()
	valid := validAlertFixture()
	valid.Channels = []platformopsapi.AlertChannel{
		{Type: "email", Config: map[string]any{}},
		{Type: "telegram", Config: map[string]any{}},
	}
	if !validAlertInput(valid) {
		t.Fatal("secret-free unique channel types were rejected")
	}
}

func TestAlertInputRejectsProviderSecrets(t *testing.T) {
	t.Parallel()
	for name, config := range map[string]map[string]any{
		"recipient": {"recipient": "operator@example.test"},
		"bot token": {"bot_token": "secret"},
		"chat id":   {"chat_id": "123"},
	} {
		name, config := name, config
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			input := validAlertFixture()
			input.Channels = []platformopsapi.AlertChannel{{
				Type: "email", Config: config,
			}}
			if validAlertInput(input) {
				t.Fatal("provider secret-bearing config was accepted")
			}
		})
	}
}

func TestAlertInputRejectsDuplicateAndExcessChannels(t *testing.T) {
	t.Parallel()
	duplicate := validAlertFixture()
	duplicate.Channels = []platformopsapi.AlertChannel{
		{Type: "email"}, {Type: "email"},
	}
	if validAlertInput(duplicate) {
		t.Fatal("duplicate channel type was accepted")
	}
	excess := validAlertFixture()
	excess.Channels = []platformopsapi.AlertChannel{
		{Type: "email"}, {Type: "telegram"}, {Type: "email"},
	}
	if validAlertInput(excess) {
		t.Fatal("more than two channel types were accepted")
	}
}

func validAlertFixture() platformopsapi.AlertRuleInput {
	return platformopsapi.AlertRuleInput{
		EnvironmentID:   "10000000-0000-4000-8000-000000000001",
		Name:            "Database degraded",
		ServiceName:     "postgres",
		Condition:       platformopsapi.AlertDegraded,
		CooldownMinutes: 5,
	}
}
