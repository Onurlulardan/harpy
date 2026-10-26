package platformops

import (
	"context"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (*httpRecoveryFixture) ReadSystemHealth(
	context.Context,
) (platformopsapi.SystemHealth, error) {
	return platformopsapi.SystemHealth{Status: platformopsapi.HealthHealthy,
		ObservedAt: time.Date(2026, 7, 23, 5, 0, 0, 0, time.UTC)}, nil
}

func (*httpRecoveryFixture) RecordTrace(
	context.Context, platformopsapi.TraceObservation,
) error {
	return nil
}

func (*httpRecoveryFixture) ListTraces(
	context.Context, string, platformopsapi.TraceFilter,
) (platformopsapi.TracePage, error) {
	return platformopsapi.TracePage{}, nil
}

func (*httpRecoveryFixture) GetTrace(
	context.Context, string, string,
) ([]platformopsapi.TraceSpan, error) {
	return nil, platformopsapi.ErrNotFound
}

func (*httpRecoveryFixture) GetTraceSpan(
	context.Context, string, string,
) (platformopsapi.TraceSpan, error) {
	return platformopsapi.TraceSpan{}, platformopsapi.ErrNotFound
}

func (*httpRecoveryFixture) ReadTraceStats(
	context.Context, string,
) (platformopsapi.TraceStats, error) {
	return platformopsapi.TraceStats{}, nil
}

func (*httpRecoveryFixture) ListAlertRules(
	context.Context, string,
) ([]platformopsapi.AlertRule, error) {
	return []platformopsapi.AlertRule{}, nil
}

func (*httpRecoveryFixture) GetAlertRule(
	context.Context, string, string,
) (platformopsapi.AlertRule, error) {
	return platformopsapi.AlertRule{}, platformopsapi.ErrNotFound
}

func (*httpRecoveryFixture) CreateAlertRule(
	context.Context, platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	return platformopsapi.AlertRule{}, nil
}

func (*httpRecoveryFixture) UpdateAlertRule(
	context.Context, string, platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	return platformopsapi.AlertRule{}, nil
}

func (*httpRecoveryFixture) DeleteAlertRule(context.Context, string, string) error {
	return nil
}

func (*httpRecoveryFixture) RecordAlertTest(
	context.Context, string, string, string,
) (platformopsapi.AlertEvent, error) {
	return platformopsapi.AlertEvent{}, nil
}

func (*httpRecoveryFixture) ListAlertEvents(
	context.Context, string, platformopsapi.AlertEventFilter,
) (platformopsapi.AlertEventPage, error) {
	return platformopsapi.AlertEventPage{}, nil
}

func (*httpRecoveryFixture) GetAlertEvent(
	context.Context, string, string,
) (platformopsapi.AlertEvent, error) {
	return platformopsapi.AlertEvent{}, platformopsapi.ErrNotFound
}
