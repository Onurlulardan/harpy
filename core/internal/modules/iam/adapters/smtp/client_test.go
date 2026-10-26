package smtp

import (
	"context"
	"errors"
	"testing"
	"time"

	gomail "github.com/wneessen/go-mail"
)

type senderFixture struct{ err error }

func (sender senderFixture) DialAndSendWithContext(
	context.Context,
	...*gomail.Msg,
) error {
	return sender.err
}

func TestClientProbeFailsClosed(t *testing.T) {
	want := errors.New("TLS unavailable")
	client, err := newClient(
		smtpSettingsFixture(t), senderFixture{},
		func(context.Context) error { return want }, time.Now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Probe(context.Background()); !errors.Is(err, want) {
		t.Fatalf("probe error = %v", err)
	}
}

func TestClientRejectsTypedNilSender(t *testing.T) {
	var sender *senderFixture
	if _, err := newClient(
		smtpSettingsFixture(t), sender,
		func(context.Context) error { return nil }, time.Now,
	); !errors.Is(err, ErrContract) {
		t.Fatalf("typed nil error = %v", err)
	}
}
