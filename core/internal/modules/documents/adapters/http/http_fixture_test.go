package documentshttp

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func routeMux(routes []Route) http.Handler {
	mux := http.NewServeMux()
	for _, route := range routes {
		mux.Handle(route.Method+" "+route.Pattern, route.Handler)
	}
	return mux
}

const testDocumentID = "018f0000-0000-7000-8000-000000000101"

type accessVerifierStub struct{}

func (accessVerifierStub) Verify(_ context.Context, request AccessRequest) (Access, error) {
	return NewAccess(uuid.MustParse("018f0000-0000-7000-8000-000000000201"),
		uuid.MustParse("018f0000-0000-7000-8000-000000000202"), 1, 2, 3,
		request.Permission, request.Operation)
}

type applicationStub func(context.Context, Request) (Response, error)

func (stub applicationStub) Dispatch(ctx context.Context, request Request) (Response, error) {
	return stub(ctx, request)
}

func newTestHandler(t *testing.T, application Application) *Handler {
	t.Helper()
	handler, err := NewHandler(Dependencies{Access: accessVerifierStub{},
		Application: application, Report: func(error) {},
		Correlation: func(*http.Request) string { return "documents-correlation" }})
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
