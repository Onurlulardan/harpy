package alertrule

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	AlertRuleCreation(context.Context, platformopsapi.AlertRuleInput) (platformopsapi.AlertRule, error)
	AlertRuleDeletion(context.Context, string, string) error
	AlertRuleListing(context.Context, string) ([]platformopsapi.AlertRule, error)
	AlertRuleLookup(context.Context, string, string) (platformopsapi.AlertRule, error)
	AlertRuleUpdate(context.Context, string, platformopsapi.AlertRuleInput) (platformopsapi.AlertRule, error)
	AlertTestStorage(context.Context, string, string, string, string) (platformopsapi.AlertEvent, error)
}

type AlertRuleOperations struct{ backend Backend }

func NewAlertRuleOperations(backend Backend) *AlertRuleOperations {
	return &AlertRuleOperations{backend: backend}
}
