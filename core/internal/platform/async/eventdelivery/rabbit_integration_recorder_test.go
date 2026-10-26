//go:build integration

package eventdelivery

import (
	"context"
	"sync"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

type integrationHandler struct {
	mutex    sync.Mutex
	seen     map[string]bool
	outcomes []brokerrabbit.HandleOutcome
	signal   chan struct{}
}

func newIntegrationHandler() *integrationHandler {
	return &integrationHandler{seen: make(map[string]bool), signal: make(chan struct{}, 8)}
}

func (handler *integrationHandler) Handle(
	_ context.Context,
	envelope eventsv1.Envelope,
	_ brokerrabbit.DeliveryMetadata,
) (brokerrabbit.HandleOutcome, error) {
	handler.mutex.Lock()
	outcome := brokerrabbit.HandleCommitted
	if handler.seen[envelope.EventID] {
		outcome = brokerrabbit.HandleDuplicate
	}
	handler.seen[envelope.EventID] = true
	handler.outcomes = append(handler.outcomes, outcome)
	handler.mutex.Unlock()
	handler.signal <- struct{}{}
	return outcome, nil
}

func (handler *integrationHandler) waitFor(count int) bool {
	deadline := time.After(90 * time.Second)
	for len(handler.snapshot()) < count {
		select {
		case <-handler.signal:
		case <-deadline:
			return false
		}
	}
	return true
}

func (handler *integrationHandler) snapshot() []brokerrabbit.HandleOutcome {
	handler.mutex.Lock()
	defer handler.mutex.Unlock()
	return append([]brokerrabbit.HandleOutcome(nil), handler.outcomes...)
}
