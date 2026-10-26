package deadletterdelivery

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type panickingConnector struct{}

func (panickingConnector) Open(context.Context) (Session, error) {
	panic("broker-secret")
}

type panickingSession struct {
	closeCall atomic.Int32
}

func (*panickingSession) Run(context.Context) error {
	panic("provider-secret")
}

func (session *panickingSession) Close() error {
	session.closeCall.Add(1)
	return nil
}

func TestManagedPanicsBecomeStableRedactedFailures(t *testing.T) {
	session := &panickingSession{}
	tests := []struct {
		name      string
		connector Connector
		closed    func() int32
		wantClose int32
	}{
		{
			name: "connector", connector: panickingConnector{},
			closed: func() int32 { return 0 }, wantClose: 0,
		},
		{
			name: "session", connector: newScriptedConnector(scriptedOpen{session: session}),
			closed: session.closeCall.Load, wantClose: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, err := New(test.connector, Settings{
				ReconnectMinimum: time.Millisecond, ReconnectMaximum: 2 * time.Millisecond,
			}, openDeadLetterAdmission())
			if err != nil {
				t.Fatal(err)
			}
			if err := runner.Start(context.Background()); err != nil {
				t.Fatal(err)
			}
			failure, open := <-runner.Failures()
			if !open || !errors.Is(failure, ErrSessionPanicked) ||
				strings.Contains(failure.Error(), "secret") {
				t.Fatalf("managed panic failure = %v, open=%t", failure, open)
			}
			if _, open := <-runner.Failures(); open {
				t.Fatal("panic failure channel remained open")
			}
			if test.closed() != test.wantClose {
				t.Fatalf("panic cleanup close count = %d", test.closed())
			}
		})
	}
}
