package smtp

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type exchangeFake struct {
	result  exchangeResult
	request exchangeRequest
	calls   int
}

func (exchange *exchangeFake) Exchange(_ context.Context, request exchangeRequest) exchangeResult {
	exchange.calls++
	exchange.request = request
	return exchange.result
}

func TestClientUsesNonSecretUsernameMetadataAndPasswordOnlySecret(t *testing.T) {
	at := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	exchange := &exchangeFake{result: exchangeResult{accepted: true,
		definitive: true, completedAt: at}}
	client, _ := NewClient(exchange, func() time.Time { return at })
	request := requestFixture(t, at)
	request.Metadata.SMTPUsername = "tenant-mailer"
	request.Credential = []byte("password-only")
	result := client.Send(t.Context(), request)
	if result.Outcome != ports.ProviderAccepted ||
		string(exchange.request.username) != "tenant-mailer" ||
		string(exchange.request.password) != "password-only" {
		t.Fatalf("result=%+v username=%q password=%q", result,
			exchange.request.username, exchange.request.password)
	}
}

func requestFixture(t *testing.T, at time.Time) ports.ProviderRequest {
	t.Helper()
	id, err := domain.ParseID("019b0000-0000-7000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	return ports.ProviderRequest{DeliveryID: id, Target: []byte("user@example.test"),
		Subject: "Task ready", Body: "Review it", TemplateKind: domain.TemplatePlainText,
		Credential: []byte("smtp-user\x00smtp-password"), Requested: at,
		Metadata: ports.ProviderMetadata{SMTPHost: "smtp.example.test", SMTPPort: 587,
			SMTPSecurity: "starttls", FromAddress: "Forge <noreply@example.test>"}}
}

func TestClientClassifiesSMTPResultsWithoutGuessingAfterBytes(t *testing.T) {
	at := time.Date(2026, 7, 18, 15, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		raw  exchangeResult
		want ports.ProviderOutcome
	}{
		{name: "accepted", raw: exchangeResult{accepted: true, definitive: true,
			code: "SMTP_ACCEPTED", class: "data_commit", completedAt: at},
			want: ports.ProviderAccepted},
		{name: "connect", raw: exchangeResult{code: "SMTP_CONNECT_FAILED",
			class: "connect", completedAt: at}, want: ports.ProviderTransient},
		{name: "authenticated temporary", raw: exchangeResult{definitive: true,
			temporary: true, code: "SMTP_451", class: "data_commit", completedAt: at},
			want: ports.ProviderTransient},
		{name: "after bytes timeout", raw: exchangeResult{bytesWritten: true,
			code: "SMTP_TRANSPORT_FAILED", class: "data_write", completedAt: at},
			want: ports.ProviderUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exchange := &exchangeFake{result: test.raw}
			client, _ := NewClient(exchange, func() time.Time { return at })
			result := client.Send(t.Context(), requestFixture(t, at))
			if result.Outcome != test.want {
				t.Fatalf("outcome = %s", result.Outcome)
			}
			if test.want == ports.ProviderAccepted &&
				result.ProviderReferenceDigest == [32]byte{} {
				t.Fatal("accepted result lacks reference digest")
			}
		})
	}
}

func TestClientRejectsHeaderInjectionBeforeNetwork(t *testing.T) {
	at := time.Date(2026, 7, 18, 15, 0, 0, 0, time.UTC)
	exchange := &exchangeFake{}
	client, _ := NewClient(exchange, func() time.Time { return at })
	request := requestFixture(t, at)
	request.Subject = "safe\r\nBcc: attacker@example.test"
	result := client.Send(t.Context(), request)
	if result.Outcome != ports.ProviderPermanent || exchange.calls != 0 {
		t.Fatalf("result = %+v, network calls = %d", result, exchange.calls)
	}
}
