package outboxdelivery

import (
	"context"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

func TestOutboxScanDoesNotTouchStoreWhileAdmissionIsClosed(t *testing.T) {
	catalog, err := eventsv1.NewProducerCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}
	gate := admission.NewGate(nil)
	runner, err := newWithJitter(
		unreachableStore{}, catalog, unavailableConnector{}, Settings{
			ConfirmTimeout: time.Second, ReconnectMinimum: time.Millisecond,
			ReconnectMaximum: time.Second,
		}, gate, fixedJitter(0),
	)
	if err != nil {
		t.Fatal(err)
	}
	worked, err := runner.scanAuthorities(context.Background(), nil)
	if err != nil || worked {
		t.Fatalf("closed admission scan worked=%t err=%v", worked, err)
	}
	if !gate.Open() {
		t.Fatal("outbox admission did not open")
	}
	if _, err := runner.scanAuthorities(context.Background(), nil); err == nil {
		t.Fatal("opened admission did not reach the test store")
	}
}
