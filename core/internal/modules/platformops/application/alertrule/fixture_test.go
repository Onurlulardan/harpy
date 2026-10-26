package alertrule

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ operation string }

func (backend *backendFixture) AlertRuleCreation(
	context.Context, platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	backend.operation = "create"
	return platformopsapi.AlertRule{ID: "rule"}, nil
}

func (backend *backendFixture) AlertRuleDeletion(context.Context, string, string) error {
	backend.operation = "delete"
	return nil
}

func (backend *backendFixture) AlertRuleListing(
	context.Context, string,
) ([]platformopsapi.AlertRule, error) {
	backend.operation = "list"
	return []platformopsapi.AlertRule{{ID: "rule"}}, nil
}

func (backend *backendFixture) AlertRuleLookup(
	context.Context, string, string,
) (platformopsapi.AlertRule, error) {
	backend.operation = "get"
	return platformopsapi.AlertRule{ID: "rule"}, nil
}

func (backend *backendFixture) AlertRuleUpdate(
	context.Context, string, platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	backend.operation = "update"
	return platformopsapi.AlertRule{ID: "rule"}, nil
}

func (backend *backendFixture) AlertTestRecording(
	context.Context, string, string, string,
) (platformopsapi.AlertEvent, error) {
	backend.operation = "record"
	return platformopsapi.AlertEvent{ID: "event"}, nil
}
