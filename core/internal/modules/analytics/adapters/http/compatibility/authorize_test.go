package compatibilityhttp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
)

var legacyNow = time.Date(2026, 7, 19, 13, 0, 0, 0, time.UTC)

type authorizeAccessStub struct{}

func (authorizeAccessStub) Verify(context.Context, httptransport.AccessRequest) (httptransport.Access, error) {
	return httptransport.Access{}, httptransport.ErrAuthentication
}

type legacyStub struct {
	key    compatibilityapp.LegacyKey
	called bool
	hit    bool
	err    error
}

func (stub *legacyStub) AuthorizeLegacy(
	_ context.Context, key compatibilityapp.LegacyKey,
) (compatibilityapp.LegacyMapping, error) {
	stub.called = true
	stub.key = key
	stub.hit = true
	if stub.err != nil {
		return compatibilityapp.LegacyMapping{}, stub.err
	}
	return compatibilityapp.LegacyMapping{Key: key, TenantID: uuid.New(),
		ReleaseID: uuid.New(), PublicID: [16]byte{1, 2, 3, 4}, State: "active",
		HitCount: 1, LastHitAt: &legacyNow, RemoveAfter: legacyNow.Add(30 * 24 * time.Hour)}, nil
}

type legacyPublicStub struct {
	called  bool
	ordered bool
	before  func() bool
	request publicationapp.PublicRequest
}

func (stub *legacyPublicStub) Execute(
	_ context.Context, request publicationapp.PublicRequest,
) (publicationapp.PublicResult, error) {
	stub.called = true
	if stub.before != nil {
		stub.ordered = stub.before()
	}
	request.Parameters = append([]byte(nil), request.Parameters...)
	stub.request = request
	return publicationapp.PublicResult{Body: []byte(`{"items":[]}`)}, nil
}

func authorizeHandler(t *testing.T) (*Handler, *legacyStub, *legacyPublicStub) {
	t.Helper()
	runtime, err := httptransport.NewRuntime(authorizeAccessStub{}, func(error) {},
		func(*http.Request) string { return "correlation-1" },
		func(*http.Request) (string, error) { return "203.0.113.10", nil }, legacyFailure)
	if err != nil {
		t.Fatal(err)
	}
	legacy := &legacyStub{}
	public := &legacyPublicStub{}
	return &Handler{runtime: runtime, application: legacy, public: public}, legacy, public
}

func legacyFailure(err error) (int, string, string) {
	if errors.Is(err, compatibilityapp.ErrInvalid) {
		return 422, "ANALYTICS_INPUT_INVALID", "Analytics input is invalid"
	}
	return 503, "ANALYTICS_UNAVAILABLE", "Analytics is unavailable"
}

func legacyRequest(method, path, body string) *http.Request {
	var reader io.Reader = http.NoBody
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.SetPathValue("slug", "orders")
	request.Header.Set("X-Tenant-Slug", "tenant-one")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}

func TestAuthorizeUsesExactMappingAndDurableHitBeforeRelease(t *testing.T) {
	handler, legacy, public := authorizeHandler(t)
	public.before = func() bool { return legacy.hit }
	request := legacyRequest(http.MethodGet,
		"/api/openapi/execute/orders?page=2&limit=25&filter[state]=open", "")
	recorder := httptest.NewRecorder()
	handler.Authorize(recorder, request)
	if recorder.Code != http.StatusOK || !legacy.called || !public.called || !public.ordered {
		t.Fatalf("status=%d legacy=%v public=%v ordered=%v",
			recorder.Code, legacy.called, public.called, public.ordered)
	}
	if legacy.key.Method != http.MethodGet || legacy.key.Path != "/api/openapi/execute/orders" ||
		legacy.key.Slug != "orders" || legacy.key.TenantIdentity != "tenant-one" ||
		public.request.ExpectedKind != publicationapp.ResourceQuery || public.request.Slug != "legacy" ||
		!strings.Contains(string(public.request.Parameters), `"filters":{"state":"open"}`) {
		t.Fatalf("key=%+v request=%+v", legacy.key, public.request)
	}
}

func TestAuthorizeRejectsHeaderAndPathSubstitutionBeforeLookup(t *testing.T) {
	mutations := []func(*http.Request){
		func(request *http.Request) { request.Header.Del("X-Tenant-Slug") },
		func(request *http.Request) { request.Header.Set("X-Tenant-ID", "tenant-id") },
		func(request *http.Request) { request.URL.Path = "/api/openapi/execute/other" },
		func(request *http.Request) { request.SetPathValue("slug", "other") },
	}
	for index, mutate := range mutations {
		handler, legacy, public := authorizeHandler(t)
		request := legacyRequest(http.MethodGet, "/api/openapi/execute/orders", "")
		mutate(request)
		recorder := httptest.NewRecorder()
		handler.Authorize(recorder, request)
		if recorder.Code != http.StatusUnprocessableEntity || legacy.called || public.called {
			t.Fatalf("case=%d status=%d legacy=%v public=%v",
				index, recorder.Code, legacy.called, public.called)
		}
	}
}

func TestAuthorizePersistsHitBeforeRejectingInvalidDTO(t *testing.T) {
	handler, legacy, public := authorizeHandler(t)
	request := legacyRequest(http.MethodPost, "/api/openapi/execute/orders", `{"unknown":true}`)
	recorder := httptest.NewRecorder()
	handler.Authorize(recorder, request)
	if recorder.Code != http.StatusUnprocessableEntity || !legacy.hit || public.called {
		t.Fatalf("status=%d hit=%v public=%v", recorder.Code, legacy.hit, public.called)
	}
}
