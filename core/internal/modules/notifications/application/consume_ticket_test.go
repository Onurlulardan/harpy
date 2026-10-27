package application

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestConsumeTicketRejectsReplay(t *testing.T) {
	application, _, _, scope := realtimeFixture(t)
	issued, err := application.IssueTicket(t.Context(), scope,
		"https://app.harpy.io")
	if err != nil {
		t.Fatal(err)
	}
	offered := domain.RealtimeProtocol + ", " + domain.TicketProtocolPrefix + issued.Token
	if _, err = application.ConsumeTicket(t.Context(), scope,
		"https://app.harpy.io", offered); err != nil {
		t.Fatalf("first ConsumeTicket() error = %v", err)
	}
	if _, err = application.ConsumeTicket(t.Context(), scope,
		"https://app.harpy.io", offered); err != ErrForbidden {
		t.Fatalf("replayed ConsumeTicket() error = %v", err)
	}
}
