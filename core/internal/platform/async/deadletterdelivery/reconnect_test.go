package deadletterdelivery

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunnerReconnectsWithinBoundsAndDrainsActiveSession(t *testing.T) {
	session := newBlockingSession()
	connector := newScriptedConnector(
		scriptedOpen{err: errors.New("offline")},
		scriptedOpen{err: errors.New("still offline")},
		scriptedOpen{session: session},
	)
	runner, err := newWithJitter(connector, Settings{
		ReconnectMinimum: time.Millisecond, ReconnectMaximum: 4 * time.Millisecond,
	}, fixedJitter(0), openDeadLetterAdmission())
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	for wanted := 1; wanted <= 3; wanted++ {
		select {
		case opened := <-connector.opened:
			if opened != wanted {
				t.Fatalf("open sequence = %d, want %d", opened, wanted)
			}
		case <-time.After(time.Second):
			t.Fatalf("connector did not reach open %d", wanted)
		}
	}
	select {
	case <-session.started:
	case <-time.After(time.Second):
		t.Fatal("reconnected session did not start")
	}
	runner.BeginDrain()
	waitContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(waitContext); err != nil {
		t.Fatal(err)
	}
	if connector.OpenCount() != 3 || session.closeCall.Load() != 1 {
		t.Fatalf("opens=%d closes=%d", connector.OpenCount(), session.closeCall.Load())
	}
	if failure, open := <-runner.Failures(); open || failure != nil {
		t.Fatalf("graceful drain failure = %v, open=%t", failure, open)
	}
}

func TestReconnectDelayUsesBoundedFullJitter(t *testing.T) {
	settings := Settings{ReconnectMinimum: time.Second, ReconnectMaximum: 30 * time.Second}
	if got := reconnectDelay(time.Second, settings, fixedJitter(99)); got != time.Second {
		t.Fatalf("minimum delay = %s", got)
	}
	if got := reconnectDelay(8*time.Second, settings, fixedJitter(^uint64(0))); got != 8*time.Second {
		t.Fatalf("maximum sampled delay = %s", got)
	}
	if got := nextReconnectCap(16*time.Second, settings.ReconnectMaximum); got != 30*time.Second {
		t.Fatalf("maximum cap = %s", got)
	}
}
