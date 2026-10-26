package eventsv1

import (
	"fmt"
	"strconv"
)

const (
	BrokerExchangeName           = "harpy.events.v1"
	DeadLetterExchangeName       = "harpy.dead_letters.v1"
	DeadLetterIngestQueue        = "harpy.dead_letters.ingest.v1"
	DeadLetterQuarantineExchange = "harpy.dead_letters.quarantine.v1"
	DeadLetterQuarantineQueue    = "harpy.dead_letters.quarantine.v1"
	DeadLetterQuarantineKey      = "invalid"
	DeadLetterQuarantineTTLMS    = int64(2_592_000_000)
	BrokerContentType            = "application/vnd.harpy.integration-event.v1+json"
)

func RoutingKey(eventType string, eventVersion int) (string, error) {
	if !validType(eventType) || eventVersion < 1 {
		return "", fmt.Errorf("event routing identity is invalid")
	}
	return eventType + ".v" + strconv.Itoa(eventVersion), nil
}

func (value Envelope) RoutingKey() (string, error) {
	return RoutingKey(value.EventType, value.EventVersion)
}

func (value ConsumerPolicy) BindingKey() (string, error) {
	if err := value.Validate(); err != nil {
		return "", err
	}
	return RoutingKey(value.EventType, value.EventVersion)
}

func (value ConsumerPolicy) DeadLetterRoutingKey() (string, error) {
	if err := value.Validate(); err != nil {
		return "", err
	}
	return "consumer." + value.QueueName, nil
}
