package publicationhttp

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

type executeAccessStub struct{}

func (executeAccessStub) Verify(context.Context, httptransport.AccessRequest) (httptransport.Access, error) {
	return httptransport.Access{}, httptransport.ErrAuthentication
}

type executePublicStub struct {
	called  bool
	request publicationapp.PublicRequest
}

func (stub *executePublicStub) Execute(
	_ context.Context, request publicationapp.PublicRequest,
) (publicationapp.PublicResult, error) {
	stub.called = true
	request.CapabilityToken = append([]byte(nil), request.CapabilityToken...)
	request.Parameters = append([]byte(nil), request.Parameters...)
	stub.request = request
	return publicationapp.PublicResult{Body: []byte(`{"items":[]}`)}, nil
}

func executeHandler(t *testing.T) (*Handler, *executePublicStub) {
	t.Helper()
	runtime, err := httptransport.NewRuntime(executeAccessStub{}, func(error) {},
		func(*http.Request) string { return "correlation-1" },
		func(*http.Request) (string, error) { return "203.0.113.10", nil },
		func(error) (int, string, string) { return 503, "ANALYTICS_UNAVAILABLE", "Analytics is unavailable" })
	if err != nil {
		t.Fatal(err)
	}
	public := &executePublicStub{}
	return &Handler{runtime: runtime, public: public}, public
}

func opaqueRequest(method, resource, rawQuery string) *http.Request {
	publicID := [16]byte{1, 2, 3, 4}
	wire := base64.RawURLEncoding.EncodeToString(publicID[:])
	path := "/api/published/" + resource + "/" + wire + "/orders"
	request := httptest.NewRequest(method, path+rawQuery, http.NoBody)
	request.SetPathValue("public_id", wire)
	request.SetPathValue("slug", "orders")
	return request
}

func TestExecuteRejectsEveryTenantHeaderBeforeResolution(t *testing.T) {
	for _, header := range []string{"X-Tenant-ID", "X-Tenant-Slug"} {
		handler, public := executeHandler(t)
		request := opaqueRequest(http.MethodGet, "queries", "")
		request.Header.Set(header, "")
		recorder := httptest.NewRecorder()
		handler.Execute(recorder, request)
		if recorder.Code != http.StatusUnprocessableEntity || public.called {
			t.Fatalf("header=%s status=%d called=%v", header, recorder.Code, public.called)
		}
	}
}

func TestExecuteUsesOnlyCanonicalAuthorizationCapability(t *testing.T) {
	handler, public := executeHandler(t)
	token := bytes.Repeat([]byte{7}, 32)
	request := opaqueRequest(http.MethodGet, "queries", "?param[state]=open&page[size]=10")
	request.Header.Set("Authorization", "Bearer "+base64.RawURLEncoding.EncodeToString(token))
	recorder := httptest.NewRecorder()
	handler.Execute(recorder, request)
	if recorder.Code != http.StatusOK || !public.called ||
		public.request.ExpectedKind != publicationapp.ResourceQuery ||
		!bytes.Equal(public.request.CapabilityToken, token) ||
		!strings.Contains(string(public.request.Parameters), `"state":"open"`) {
		t.Fatalf("status=%d request=%+v", recorder.Code, public.request)
	}
}

func TestExecuteRejectsCapabilityAndResourceSubstitution(t *testing.T) {
	handler, public := executeHandler(t)
	request := opaqueRequest(http.MethodGet, "dashboards", "")
	request.Header.Set("Authorization", "Bearer not-canonical")
	recorder := httptest.NewRecorder()
	handler.Execute(recorder, request)
	if recorder.Code != http.StatusForbidden || public.called {
		t.Fatalf("status=%d called=%v", recorder.Code, public.called)
	}
	handler, public = executeHandler(t)
	request = opaqueRequest(http.MethodGet, "dashboards", "")
	recorder = httptest.NewRecorder()
	handler.Execute(recorder, request)
	if recorder.Code != http.StatusOK || public.request.ExpectedKind != publicationapp.ResourceDashboard {
		t.Fatalf("status=%d kind=%s", recorder.Code, public.request.ExpectedKind)
	}
}

func TestExecutePOSTRejectsUnknownWireMembers(t *testing.T) {
	handler, public := executeHandler(t)
	request := opaqueRequest(http.MethodPost, "queries", "")
	request.Body = ioNopCloser(`{"params":{},"unknown":true}`)
	request.ContentLength = int64(len(`{"params":{},"unknown":true}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.Execute(recorder, request)
	if recorder.Code != http.StatusUnprocessableEntity || public.called {
		t.Fatalf("status=%d called=%v", recorder.Code, public.called)
	}
}

func ioNopCloser(value string) *readCloser { return &readCloser{Reader: strings.NewReader(value)} }

type readCloser struct{ *strings.Reader }

func (reader *readCloser) Close() error { return nil }
