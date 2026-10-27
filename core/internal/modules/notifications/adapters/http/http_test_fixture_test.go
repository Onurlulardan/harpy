package notificationhttp

import (
	"context"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/google/uuid"
)

var (
	httpTestTenant = uuid.MustParse("10000000-0000-4000-8000-000000000001")
	httpTestUser   = uuid.MustParse("10000000-0000-4000-8000-000000000002")
)

type accessVerifierFake struct {
	request AccessRequest
	access  Access
	calls   int
}

func (fake *accessVerifierFake) Verify(_ context.Context, request AccessRequest) (Access, error) {
	fake.request, fake.calls = request, fake.calls+1
	return fake.access, nil
}

type httpApplicationFake struct {
	ticketCalls, consumeCalls int
	lastRealtime              RealtimeInput
	lastIntent                IntentInput
}

func (*httpApplicationFake) ListMessages(context.Context, ListInput) (ports.MessagePage, error) {
	return ports.MessagePage{}, nil
}
func (*httpApplicationFake) GetMessage(context.Context, MessageInput) (domain.Message, error) {
	return domain.Message{}, nil
}
func (*httpApplicationFake) MarkRead(context.Context, MessageInput) (domain.Message, error) {
	return domain.Message{}, nil
}
func (*httpApplicationFake) MarkAllRead(context.Context, Access) (int64, error) { return 0, nil }
func (*httpApplicationFake) DismissMessage(context.Context, MessageInput) (domain.Message, error) {
	return domain.Message{}, nil
}
func (*httpApplicationFake) GetPreferences(context.Context, Access) ([]domain.Preference, error) {
	return nil, nil
}
func (*httpApplicationFake) UpdatePreferences(context.Context, PreferencesInput) ([]domain.Preference, error) {
	return nil, nil
}
func (fake *httpApplicationFake) RecordIntent(_ context.Context,
	input IntentInput,
) (api.IntentResult, error) {
	fake.lastIntent = input
	return api.IntentResult{}, nil
}
func (fake *httpApplicationFake) CreateRealtimeTicket(context.Context, Access) (application.RealtimeTicketResult, error) {
	fake.ticketCalls++
	return application.RealtimeTicketResult{Token: "token"}, nil
}
func (fake *httpApplicationFake) ConsumeRealtimeTicket(_ context.Context,
	input RealtimeInput,
) (application.RealtimeConnectionIdentity, error) {
	fake.consumeCalls++
	fake.lastRealtime = input
	tenant, _ := domain.IDFromBytes(httpTestTenant[:])
	user, _ := domain.IDFromBytes(httpTestUser[:])
	return application.RealtimeConnectionIdentity{TenantID: tenant, UserID: user}, nil
}

type realtimeFake struct{ serveCalls int }

func (fake *realtimeFake) Serve(http.ResponseWriter, *http.Request,
	application.RealtimeConnectionIdentity,
) error {
	fake.serveCalls++
	return nil
}
func (*realtimeFake) Broadcast(context.Context, RealtimeHint) error { return nil }

func newHTTPTestHandler(access Access, app *httpApplicationFake) *Handler {
	handler, _ := NewHandler(HandlerDependencies{Access: &accessVerifierFake{access: access},
		Application: app, Realtime: &realtimeFake{}, Report: func(error) {},
		Correlation: func(*http.Request) string { return "10000000-0000-4000-8000-000000000099" }})
	return handler
}
