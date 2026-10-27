package platformops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

const (
	httpAlertRuleID  = "018f0000-0000-7000-8000-000000000101"
	httpAlertEventID = "018f0000-0000-7000-8000-000000000102"
	httpTraceSpanID  = "0123456789abcdef"
)

type httpDetailFixture struct{ *httpRecoveryFixture }

func (*httpDetailFixture) GetAlertRule(
	context.Context, string, string,
) (platformopsapi.AlertRule, error) {
	return platformopsapi.AlertRule{ID: httpAlertRuleID, Name: "availability",
		Description: "", Enabled: true, Condition: platformopsapi.AlertUnhealthy,
		CooldownMinutes: 5, Channels: []platformopsapi.AlertChannel{},
		CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(1, 0).UTC()}, nil
}

func (*httpDetailFixture) GetAlertEvent(
	context.Context, string, string,
) (platformopsapi.AlertEvent, error) {
	return platformopsapi.AlertEvent{ID: httpAlertEventID, ServiceName: "core-api",
		PreviousStatus: platformopsapi.HealthHealthy,
		CurrentStatus:  platformopsapi.HealthUnhealthy, ChannelsUsed: []string{},
		Message: "probe", CreatedAt: time.Unix(1, 0).UTC()}, nil
}

func (*httpDetailFixture) GetTraceSpan(
	context.Context, string, string,
) (platformopsapi.TraceSpan, error) {
	return platformopsapi.TraceSpan{TraceObservation: platformopsapi.TraceObservation{
		CorrelationID: "request-1", SpanID: httpTraceSpanID, ServiceName: "core-api",
		Operation: "read", SpanKind: "server", Method: http.MethodGet,
		Path: "/api/example", StatusCode: http.StatusOK,
		StartedAt: time.Unix(1, 0).UTC(), EndedAt: time.Unix(2, 0).UTC(),
		DurationMillis: 1000, Level: "info"}, CreatedAt: time.Unix(2, 0).UTC()}, nil
}

func TestHTTPDetailReadsReturnOwnedResources(t *testing.T) {
	handler, _, _ := httpHandlerFixture()
	fixture := &httpDetailFixture{httpRecoveryFixture: &httpRecoveryFixture{}}
	handler.alerts, handler.traces = fixture, fixture
	cases := []struct {
		path, key, value, expected string
		serve                      func(http.ResponseWriter, *http.Request)
	}{
		{"/api/platform/operations/alert-rules/" + httpAlertRuleID,
			"alert_rule_id", httpAlertRuleID, httpAlertRuleID, handler.getAlertRule},
		{"/api/platform/operations/alert-events/" + httpAlertEventID,
			"alert_event_id", httpAlertEventID, httpAlertEventID, handler.getAlertEvent},
		{"/api/platform/operations/traces/spans/" + httpTraceSpanID,
			"span_id", httpTraceSpanID, httpTraceSpanID, handler.getTraceSpan},
	}
	for _, test := range cases {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, test.path, nil)
		request.SetPathValue(test.key, test.value)
		test.serve(recorder, request)
		if recorder.Code != http.StatusOK ||
			!strings.Contains(recorder.Body.String(), test.expected) {
			t.Fatalf("GET %s = %d %s", test.path, recorder.Code, recorder.Body.String())
		}
	}
}
