package ports

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type TraceStore interface {
	RecordTrace(context.Context, platformopsapi.TraceObservation) error
	ListTraces(context.Context, string, platformopsapi.TraceFilter) (platformopsapi.TracePage, error)
	GetTrace(context.Context, string, string) ([]platformopsapi.TraceSpan, error)
	GetTraceSpan(context.Context, string, string) (platformopsapi.TraceSpan, error)
	ReadTraceStats(context.Context, string) (platformopsapi.TraceStats, error)
}

type AlertStore interface {
	ListAlertRules(context.Context, string) ([]platformopsapi.AlertRule, error)
	GetAlertRule(context.Context, string, string) (platformopsapi.AlertRule, error)
	CreateAlertRule(context.Context, platformopsapi.AlertRuleInput) (platformopsapi.AlertRule, error)
	UpdateAlertRule(context.Context, string, platformopsapi.AlertRuleInput) (platformopsapi.AlertRule, error)
	DeleteAlertRule(context.Context, string, string) error
	RecordAlertTest(context.Context, string, string, string, string) (platformopsapi.AlertEvent, error)
	ListAlertEvents(context.Context, string, platformopsapi.AlertEventFilter) (platformopsapi.AlertEventPage, error)
	GetAlertEvent(context.Context, string, string) (platformopsapi.AlertEvent, error)
}
