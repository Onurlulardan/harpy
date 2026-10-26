package telegram

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type doerFake struct {
	status     int
	body       string
	err        error
	wrote      bool
	calls      int
	requestURL string
}

func (doer *doerFake) Do(request *http.Request) (*http.Response, error) {
	doer.calls++
	doer.requestURL = request.URL.Redacted()
	if doer.wrote {
		if trace := httptrace.ContextClientTrace(request.Context()); trace != nil {
			trace.WroteRequest(httptrace.WroteRequestInfo{})
		}
	}
	if doer.err != nil {
		return nil, doer.err
	}
	return &http.Response{StatusCode: doer.status,
		Body: io.NopCloser(strings.NewReader(doer.body)), Header: make(http.Header)}, nil
}

func telegramRequest(t *testing.T, at time.Time) ports.ProviderRequest {
	t.Helper()
	id, err := domain.ParseID("019b0000-0000-7000-8000-000000000002")
	if err != nil {
		t.Fatal(err)
	}
	return ports.ProviderRequest{DeliveryID: id, Target: []byte("-1001234567890"),
		Body: "Task ready", Subject: "Task", TemplateKind: domain.TemplatePlainText,
		Credential: []byte("123456789:abcdefghijklmnopqrstuvwxyz_ABCD"), Requested: at,
		Metadata: ports.ProviderMetadata{TelegramAPIBase: "https://api.telegram.org"}}
}

func TestTelegramClassifiesAuthenticatedResponses(t *testing.T) {
	at := time.Date(2026, 7, 18, 16, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		status int
		body   string
		want   ports.ProviderOutcome
	}{
		{name: "accepted", status: 200,
			body: `{"ok":true,"result":{"message_id":42}}`, want: ports.ProviderAccepted},
		{name: "rate limited", status: 429, body: `{}`, want: ports.ProviderTransient},
		{name: "provider failure", status: 500, body: `{}`, want: ports.ProviderUnknown},
		{name: "rejected", status: 403, body: `{}`, want: ports.ProviderPermanent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doer := &doerFake{status: test.status, body: test.body}
			client, _ := NewClient(doer, func() time.Time { return at })
			result := client.Send(t.Context(), telegramRequest(t, at))
			if result.Outcome != test.want {
				t.Fatalf("result = %+v", result)
			}
			if test.want == ports.ProviderAccepted &&
				result.ProviderReferenceDigest == [32]byte{} {
				t.Fatal("accepted response has no reference digest")
			}
		})
	}
}

func TestTelegramNetworkFailureUsesWriteBoundary(t *testing.T) {
	at := time.Date(2026, 7, 18, 16, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		wrote bool
		want  ports.ProviderOutcome
	}{{false, ports.ProviderTransient}, {true, ports.ProviderUnknown}} {
		doer := &doerFake{err: errors.New("network unavailable"), wrote: test.wrote}
		client, _ := NewClient(doer, func() time.Time { return at })
		result := client.Send(t.Context(), telegramRequest(t, at))
		if result.Outcome != test.want {
			t.Fatalf("wrote=%t result=%+v", test.wrote, result)
		}
	}
}

func TestTelegramRejectsNonCanonicalEndpointBeforeNetwork(t *testing.T) {
	at := time.Date(2026, 7, 18, 16, 0, 0, 0, time.UTC)
	doer := &doerFake{}
	client, _ := NewClient(doer, func() time.Time { return at })
	request := telegramRequest(t, at)
	request.Metadata.TelegramAPIBase = "https://evil.example"
	result := client.Send(t.Context(), request)
	if result.Outcome != ports.ProviderPermanent || doer.calls != 0 {
		t.Fatalf("result=%+v calls=%d", result, doer.calls)
	}
}
