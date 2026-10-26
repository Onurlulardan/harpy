package async

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

type EventCatalogs struct {
	producers eventsv1.ProducerCatalog
	consumers eventsv1.ConsumerCatalog
}

func NewEventCatalogs(
	producerPolicies []eventsv1.ProducerPolicy,
	consumerPolicies []eventsv1.ConsumerPolicy,
) (EventCatalogs, error) {
	producers, err := eventsv1.NewProducerCatalog(producerPolicies)
	if err != nil {
		return EventCatalogs{}, fmt.Errorf("build event producer catalog: %w", err)
	}
	consumers, err := eventsv1.NewConsumerCatalog(consumerPolicies)
	if err != nil {
		return EventCatalogs{}, fmt.Errorf("build event consumer catalog: %w", err)
	}
	return EventCatalogs{producers: producers, consumers: consumers}, nil
}

func (catalogs EventCatalogs) Producers() eventsv1.ProducerCatalog {
	copy, _ := eventsv1.NewProducerCatalog(catalogs.producers.Policies())
	return copy
}

func (catalogs EventCatalogs) Consumers() eventsv1.ConsumerCatalog {
	copy, _ := eventsv1.NewConsumerCatalog(catalogs.consumers.Policies())
	return copy
}
