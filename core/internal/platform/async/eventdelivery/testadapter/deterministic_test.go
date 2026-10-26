package testadapter

import (
	"context"
	"errors"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

func TestDeterministicHandlerCapturesDefensiveCallEvidence(t *testing.T) {
	handler, err := New(brokerrabbit.HandleDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := handler.Handle(
		context.Background(), eventsv1.Envelope{EventID: "event-1"},
		brokerrabbit.DeliveryMetadata{DeliveryCount: 2, Redelivered: true},
	)
	if err != nil || outcome != brokerrabbit.HandleDuplicate {
		t.Fatalf("unexpected deterministic outcome: %s %v", outcome, err)
	}
	calls := handler.Calls()
	if len(calls) != 1 || calls[0].EventID != "event-1" ||
		calls[0].DeliveryCount != 2 || !calls[0].Redelivered {
		t.Fatalf("unexpected captured call: %+v", calls)
	}
	calls[0].EventID = "mutated"
	if handler.Calls()[0].EventID == "mutated" {
		t.Fatal("test handler exposed mutable call storage")
	}
}

func TestDeterministicHandlerRejectsUnknownOutcomeAndCancellation(t *testing.T) {
	if _, err := New("unknown"); !errors.Is(err, ErrOutcomeInvalid) {
		t.Fatalf("unknown outcome was accepted: %v", err)
	}
	handler, _ := New(brokerrabbit.HandleCommitted)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	outcome, err := handler.Handle(ctx, eventsv1.Envelope{}, brokerrabbit.DeliveryMetadata{})
	if !errors.Is(err, context.Canceled) || outcome != brokerrabbit.HandleTransient {
		t.Fatalf("cancellation did not remain transient: %s %v", outcome, err)
	}
}
