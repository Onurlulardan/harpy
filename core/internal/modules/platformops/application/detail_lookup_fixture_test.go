package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

const (
	detailEnvironmentID = "018f0000-0000-7000-8000-000000000201"
	detailRuleID        = "018f0000-0000-7000-8000-000000000202"
	detailEventID       = "018f0000-0000-7000-8000-000000000203"
	detailSpanID        = "0123456789abcdef"
)

type detailLookupFixture struct{ operation string }

func (*detailLookupFixture) RecordTrace(context.Context, platformopsapi.TraceObservation) error {
	return nil
}
func (*detailLookupFixture) ListTraces(
	context.Context, string, platformopsapi.TraceFilter,
) (platformopsapi.TracePage, error) {
	return platformopsapi.TracePage{}, nil
}
func (*detailLookupFixture) GetTrace(
	context.Context, string, string,
) ([]platformopsapi.TraceSpan, error) {
	return nil, nil
}
func (fixture *detailLookupFixture) GetTraceSpan(
	context.Context, string, string,
) (platformopsapi.TraceSpan, error) {
	fixture.operation = "trace_span"
	return platformopsapi.TraceSpan{TraceObservation: platformopsapi.TraceObservation{
		SpanID: detailSpanID}}, nil
}
func (*detailLookupFixture) ReadTraceStats(
	context.Context, string,
) (platformopsapi.TraceStats, error) {
	return platformopsapi.TraceStats{}, nil
}
func (*detailLookupFixture) ListAlertRules(
	context.Context, string,
) ([]platformopsapi.AlertRule, error) {
	return nil, nil
}
func (fixture *detailLookupFixture) GetAlertRule(
	context.Context, string, string,
) (platformopsapi.AlertRule, error) {
	fixture.operation = "alert_rule"
	return platformopsapi.AlertRule{ID: detailRuleID}, nil
}
func (*detailLookupFixture) CreateAlertRule(
	context.Context, platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	return platformopsapi.AlertRule{}, nil
}
func (*detailLookupFixture) UpdateAlertRule(
	context.Context, string, platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	return platformopsapi.AlertRule{}, nil
}
func (*detailLookupFixture) DeleteAlertRule(context.Context, string, string) error { return nil }
func (*detailLookupFixture) RecordAlertTest(
	context.Context, string, string, string, string,
) (platformopsapi.AlertEvent, error) {
	return platformopsapi.AlertEvent{}, nil
}
func (*detailLookupFixture) ListAlertEvents(
	context.Context, string, platformopsapi.AlertEventFilter,
) (platformopsapi.AlertEventPage, error) {
	return platformopsapi.AlertEventPage{}, nil
}
func (fixture *detailLookupFixture) GetAlertEvent(
	context.Context, string, string,
) (platformopsapi.AlertEvent, error) {
	fixture.operation = "alert_event"
	return platformopsapi.AlertEvent{ID: detailEventID}, nil
}
