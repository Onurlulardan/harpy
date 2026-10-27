package application

import (
	"bytes"
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type ticketStoreFake struct {
	ticket domain.Ticket
	hit    ports.LegacyConnectionHit
}

func (store *ticketStoreFake) BindRealtime(context.Context,
	api.Scope,
) (ports.RealtimeTicketStore, error) {
	return store, nil
}

func (store *ticketStoreFake) InsertTicket(_ context.Context, ticket domain.Ticket) error {
	store.ticket = ticket
	return nil
}

func (store *ticketStoreFake) ConsumeTicket(_ context.Context,
	input domain.TicketConsume,
) (domain.Ticket, error) {
	consumed, err := store.ticket.Consume(input)
	if err == nil {
		store.ticket = consumed
	}
	return consumed, err
}

func (store *ticketStoreFake) BindLegacyConnection(context.Context,
	api.Scope,
) (ports.LegacyConnectionStore, error) {
	return store, nil
}

func (store *ticketStoreFake) InsertLegacyConnectionHit(_ context.Context,
	hit ports.LegacyConnectionHit,
) error {
	store.hit = hit
	return nil
}

func realtimeFixture(t *testing.T) (*RealtimeApplication, *Application,
	*ticketStoreFake, testScope,
) {
	t.Helper()
	base, _, scope, _ := newIntentFixture(t)
	store := &ticketStoreFake{}
	application, err := NewRealtimeApplication(base, store,
		bytes.NewReader(bytes.Repeat([]byte{7}, 96)))
	if err != nil {
		t.Fatal(err)
	}
	return application, base, store, scope
}

func TestIssueTicketBindsActorOriginAndSingleUse(t *testing.T) {
	application, _, _, scope := realtimeFixture(t)
	created, err := application.IssueTicket(t.Context(), scope,
		"https://app.harpy.io")
	if err != nil || created.Token == "" {
		t.Fatalf("IssueTicket() = %+v, %v", created, err)
	}
	offered := domain.RealtimeProtocol + ", " + domain.TicketProtocolPrefix + created.Token
	identity, err := application.ConsumeTicket(t.Context(), scope,
		"https://app.harpy.io", offered)
	if err != nil || identity.UserID.String() == "" {
		t.Fatalf("ConsumeTicket() = %+v, %v", identity, err)
	}
	if _, err = application.ConsumeTicket(t.Context(), scope,
		"https://app.harpy.io", offered); err != ErrForbidden {
		t.Fatalf("replay error = %v", err)
	}
}

func TestIssueTicketRejectsOriginAndAuthorizationVersionChange(t *testing.T) {
	application, base, _, scope := realtimeFixture(t)
	created, err := application.IssueTicket(t.Context(), scope,
		"https://app.harpy.io")
	if err != nil {
		t.Fatal(err)
	}
	offered := domain.RealtimeProtocol + ", " + domain.TicketProtocolPrefix + created.Token
	if _, err = application.ConsumeTicket(t.Context(), scope,
		"https://evil.example", offered); err != ErrForbidden {
		t.Fatalf("origin mismatch error = %v", err)
	}
	current := base.recipients.(testRecipients)
	current.value.AuthorizationVersion++
	base.recipients = current
	if _, err = application.ConsumeTicket(t.Context(), scope,
		"https://app.harpy.io", offered); err != ErrForbidden {
		t.Fatalf("authorization version error = %v", err)
	}
}
