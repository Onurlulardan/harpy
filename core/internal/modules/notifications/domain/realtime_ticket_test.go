package domain

import (
	"bytes"
	"testing"
	"time"
)

func ticketFixture(t *testing.T) (Ticket, string, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	id, _ := NewV7(now, bytes.NewReader(bytes.Repeat([]byte{1}, 10)))
	tenant, _ := ParseID("11111111-1111-4111-8111-111111111111")
	user, _ := ParseID("22222222-2222-4222-8222-222222222222")
	ticket, token, err := NewTicket(id, tenant, user, 7, "https://app.harpy.io",
		now, bytes.NewReader(bytes.Repeat([]byte{9}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	return ticket, token, now
}

func TestTicketRequiresOriginAuthVersionAndSingleUse(t *testing.T) {
	ticket, token, now := ticketFixture(t)
	input := TicketConsume{TenantID: ticket.TenantID, UserID: ticket.UserID,
		AuthorizationVersion: 7, Origin: ticket.Origin, Token: token, At: now.Add(time.Second)}
	consumed, err := ticket.Consume(input)
	if err != nil || consumed.ConsumedAt == nil {
		t.Fatalf("Consume() = %+v, %v", consumed, err)
	}
	if _, err = consumed.Consume(input); err != ErrTicketConflict {
		t.Fatalf("ticket replay error = %v", err)
	}
	input.AuthorizationVersion++
	if _, err = ticket.Consume(input); err != ErrTicketConflict {
		t.Fatalf("stale authorization error = %v", err)
	}
}

func TestTicketProtocolNeverEchoesTicketValue(t *testing.T) {
	_, token, _ := ticketFixture(t)
	offered := RealtimeProtocol + ", " + TicketProtocolPrefix + token
	parsed, err := ParseOfferedProtocols(offered)
	if err != nil || parsed != token {
		t.Fatalf("ParseOfferedProtocols() = %q, %v", parsed, err)
	}
	if RealtimeProtocol == TicketProtocolPrefix+token {
		t.Fatal("server response protocol contains the ticket")
	}
}
