package rabbitmq

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

const ConsumerPrefetch = 32

type HandleOutcome string

const (
	HandleCommitted HandleOutcome = "committed"
	HandleDuplicate HandleOutcome = "duplicate"
	HandleTransient HandleOutcome = "transient"
	HandlePermanent HandleOutcome = "permanent"
)

type DeliveryMetadata struct {
	ConsumerPolicy eventsv1.ConsumerPolicy
	DeliveryCount  int
	Redelivered    bool
}

type Handler interface {
	Handle(context.Context, eventsv1.Envelope, DeliveryMetadata) (HandleOutcome, error)
}

func (outcome HandleOutcome) valid() bool {
	switch outcome {
	case HandleCommitted, HandleDuplicate, HandleTransient, HandlePermanent:
		return true
	default:
		return false
	}
}

func consumerTag(consumer string, version int) (string, error) {
	if version < 1 {
		return "", fmt.Errorf("consumer version is invalid")
	}
	return fmt.Sprintf("harpy.%s.v%d", consumer, version), nil
}
