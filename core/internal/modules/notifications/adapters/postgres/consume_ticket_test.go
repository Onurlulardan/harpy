package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestConsumeTicketRejectsInvalidTicketProof(t *testing.T) {
	if _, err := (&Store{}).ConsumeTicket(t.Context(), domain.TicketConsume{}); err != ErrInvalid {
		t.Fatalf("ConsumeTicket() error = %v", err)
	}
}
