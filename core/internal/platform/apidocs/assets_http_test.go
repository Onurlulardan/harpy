package apidocs

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestHandlerServesExactLocalAssetsForGetAndHead(t *testing.T) {
	handler, err := New(true, []byte(testDocument))
	if err != nil {
		t.Fatal(err)
	}
	routes := []struct {
		path        string
		contentType string
		body        []byte
	}{
		{"/swagger/", "text/html; charset=utf-8", indexHTML},
		{"/swagger/index.html", "text/html; charset=utf-8", indexHTML},
		{"/swagger/app.js", "text/javascript; charset=utf-8", applicationJS},
		{"/swagger/style.css", "text/css; charset=utf-8", stylesheet},
		{"/swagger/openapi.json", "application/vnd.oai.openapi+json", []byte(testDocument)},
	}
	for _, route := range routes {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			name := method + " " + route.path
			t.Run(name, func(t *testing.T) {
				response := request(t, handler, method, route.path)
				if response.Code != http.StatusOK {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
				if got := response.Header().Get("Content-Type"); got != route.contentType {
					t.Fatalf("content-type=%q want=%q", got, route.contentType)
				}
				if got := response.Header().Get("Content-Length"); got != strconv.Itoa(len(route.body)) {
					t.Fatalf("content-length=%q want=%d", got, len(route.body))
				}
				assertBoundaryHeaders(t, response)
				if method == http.MethodHead && response.Body.Len() != 0 {
					t.Fatalf("HEAD body=%q", response.Body.String())
				}
				if method == http.MethodGet && response.Body.String() != string(route.body) {
					t.Fatal("GET response did not preserve embedded asset bytes")
				}
			})
		}
	}
}

func TestDocumentationAssetsAreSelfContained(t *testing.T) {
	bundle := strings.ToLower(string(indexHTML) + string(applicationJS) + string(stylesheet))
	for _, forbidden := range []string{
		"http://", "https://", "//cdn", "unpkg", "jsdelivr", "cdnjs",
		"localstorage", "sessionstorage", "authorization", "bearer ",
	} {
		if strings.Contains(bundle, forbidden) {
			t.Fatalf("documentation assets reference forbidden external source %q", forbidden)
		}
	}
	for _, required := range []string{
		`href="/swagger/style.css"`,
		`src="/swagger/app.js"`,
		`href="/swagger/openapi.json"`,
		`fetch("/swagger/openapi.json"`,
		`"trace"`,
	} {
		if !strings.Contains(string(indexHTML)+string(applicationJS), required) {
			t.Fatalf("local documentation contract %q is missing", required)
		}
	}
}

func TestHandlerRedirectAndConditionalResponseRetainBoundaryHeaders(t *testing.T) {
	handler, err := New(true, []byte(testDocument))
	if err != nil {
		t.Fatal(err)
	}
	redirect := request(t, handler, http.MethodGet, "/swagger")
	if redirect.Code != http.StatusPermanentRedirect || redirect.Header().Get("Location") != "/swagger/" {
		t.Fatalf("redirect status=%d location=%q", redirect.Code, redirect.Header().Get("Location"))
	}
	assertBoundaryHeaders(t, redirect)

	initial := request(t, handler, http.MethodGet, "/swagger/openapi.json")
	conditionalRequest := httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil)
	conditionalRequest.Header.Set("If-None-Match", initial.Header().Get("ETag"))
	conditional := httptest.NewRecorder()
	handler.ServeHTTP(conditional, conditionalRequest)
	if conditional.Code != http.StatusNotModified || conditional.Body.Len() != 0 {
		t.Fatalf("conditional status=%d body=%q", conditional.Code, conditional.Body.String())
	}
	if conditional.Header().Get("ETag") != initial.Header().Get("ETag") {
		t.Fatal("conditional response omitted the document ETag")
	}
	assertBoundaryHeaders(t, conditional)
}

func assertBoundaryHeaders(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	header := response.Header()
	if header.Get("Cache-Control") != "no-store" || header.Get("Referrer-Policy") != "no-referrer" ||
		header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("boundary headers are incomplete: %v", header)
	}
	wantCSP := "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; " +
		"base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
	if csp := header.Get("Content-Security-Policy"); csp != wantCSP {
		t.Fatalf("content security policy=%q want=%q", csp, wantCSP)
	}
}
