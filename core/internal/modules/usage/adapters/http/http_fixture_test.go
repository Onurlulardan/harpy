package usagehttp

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

var (
	httpActor       = uuid.MustParse("10000000-0000-4000-8000-000000000001")
	httpTenant      = uuid.MustParse("20000000-0000-4000-8000-000000000002")
	httpEnvironment = uuid.MustParse("30000000-0000-4000-8000-000000000003")
)

type verifierStub struct {
	last AccessRequest
}

func (verifier *verifierStub) Verify(_ context.Context, request AccessRequest) (application.Access, error) {
	verifier.last = request
	input := application.AccessInput{ActorID: httpActor, SubjectTenantID: httpTenant,
		EnvironmentID: httpEnvironment, TokenVersion: 1, AuthorizationVersion: 2,
		AuthorityClock: 3, Scope: request.Scope, Permission: request.Permission,
		Operation: request.Operation, Mode: application.ModeSaaS}
	if request.Scope == domain.TenantScope {
		input.TargetTenantID = httpTenant
		input.Purpose = application.PurposeSaaSCustomer
	} else {
		input.Purpose = application.PurposeSaaSControl
		input.ControlPrimary = true
	}
	return application.NewAccess(input)
}

type applicationStub struct {
	overviewCalls int
	summaryUser   string
}

func (stub *applicationStub) Overview(context.Context, application.Access,
	application.TimeRangeRequest,
) (domain.Overview, error) {
	stub.overviewCalls++
	return domain.Overview{}, nil
}
func (*applicationStub) ActiveUsers(context.Context, application.Access) (domain.ActiveUsers, error) {
	return domain.ActiveUsers{}, nil
}
func (*applicationStub) TopUsers(context.Context, application.Access,
	application.TimeRangeRequest, int,
) ([]domain.TopUser, error) {
	return nil, nil
}
func (stub *applicationStub) UserSummary(_ context.Context, _ application.Access,
	user string, _ application.TimeRangeRequest,
) (domain.UserSummary, error) {
	stub.summaryUser = user
	return domain.UserSummary{SubjectID: user,
		LastActive: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)}, nil
}
func (*applicationStub) UserTimeline(context.Context, application.Access, string,
	int, int,
) (domain.TimelinePage, error) {
	return domain.TimelinePage{}, nil
}
func (*applicationStub) Features(context.Context, application.Access,
	application.TimeRangeRequest, int,
) ([]domain.FeatureUsage, error) {
	return nil, nil
}
func (*applicationStub) Heatmap(context.Context, application.Access,
	application.TimeRangeRequest,
) ([]domain.HeatmapCell, error) {
	return nil, nil
}
func (*applicationStub) Trend(context.Context, application.Access,
	application.TimeRangeRequest, domain.Granularity,
) ([]domain.TrendPoint, error) {
	return nil, nil
}

func testHandler(t *testing.T) (*Handler, *verifierStub, *applicationStub) {
	t.Helper()
	verifier, app := &verifierStub{}, &applicationStub{}
	handler, err := NewHandler(Dependencies{Access: verifier, Application: app,
		Report: func(error) {}, Correlation: func(*http.Request) string { return "correlation" }})
	if err != nil {
		t.Fatal(err)
	}
	return handler, verifier, app
}

func testMux(t *testing.T, handler *Handler) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	for _, route := range handler.Routes() {
		mux.Handle(route.Method+" "+route.Pattern, route.Handler)
	}
	return mux
}
