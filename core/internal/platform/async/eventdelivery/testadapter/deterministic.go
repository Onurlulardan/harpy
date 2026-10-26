package testadapter

import (
	"context"
	"errors"
	"sync"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

var ErrOutcomeInvalid = errors.New("deterministic event handler outcome is invalid")

type Call struct {
	EventID       string
	DeliveryCount int
	Redelivered   bool
}

type Handler struct {
	mutex   sync.Mutex
	outcome brokerrabbit.HandleOutcome
	calls   []Call
}

func New(outcome brokerrabbit.HandleOutcome) (*Handler, error) {
	switch outcome {
	case brokerrabbit.HandleCommitted, brokerrabbit.HandleDuplicate,
		brokerrabbit.HandleTransient, brokerrabbit.HandlePermanent:
		return &Handler{outcome: outcome}, nil
	default:
		return nil, ErrOutcomeInvalid
	}
}

func (handler *Handler) Handle(
	ctx context.Context,
	envelope eventsv1.Envelope,
	metadata brokerrabbit.DeliveryMetadata,
) (brokerrabbit.HandleOutcome, error) {
	if err := ctx.Err(); err != nil {
		return brokerrabbit.HandleTransient, err
	}
	handler.mutex.Lock()
	defer handler.mutex.Unlock()
	handler.calls = append(handler.calls, Call{
		EventID: envelope.EventID, DeliveryCount: metadata.DeliveryCount,
		Redelivered: metadata.Redelivered,
	})
	return handler.outcome, nil
}

func (handler *Handler) Calls() []Call {
	handler.mutex.Lock()
	defer handler.mutex.Unlock()
	return append([]Call(nil), handler.calls...)
}
