package rabbitmq

import (
	"fmt"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	amqp "github.com/rabbitmq/amqp091-go"
)

func deadLetterEvidence(headers amqp.Table) (deadletter.DeliveryEvidence, error) {
	reason, ok := headerString(headers, "x-first-death-reason")
	count := 0
	if deaths, exists := headers["x-death"]; exists {
		death, found := firstDeath(deaths)
		if !found {
			return deadletter.DeliveryEvidence{}, fmt.Errorf("RabbitMQ x-death header is invalid")
		}
		if !ok {
			reason, ok = headerString(death, "reason")
		}
		value, countOK := headerInteger(death, "count")
		if !countOK || value < 0 {
			return deadletter.DeliveryEvidence{}, fmt.Errorf("RabbitMQ x-death count is invalid")
		}
		count = value
	}
	if !ok {
		return deadletter.DeliveryEvidence{}, fmt.Errorf("RabbitMQ dead-letter reason is missing")
	}
	class := deadletter.ReasonPermanent
	switch reason {
	case "delivery_limit", "expired", "maxlen", "maxlen-bytes":
		class = deadletter.ReasonDeliveryLimit
	case "rejected":
	default:
		return deadletter.DeliveryEvidence{}, fmt.Errorf("RabbitMQ dead-letter reason is unsupported")
	}
	return deadletter.DeliveryEvidence{Reason: class, DeliveryCount: count}, nil
}

func firstDeath(value any) (amqp.Table, bool) {
	switch deaths := value.(type) {
	case []any:
		if len(deaths) == 0 {
			return nil, false
		}
		death, ok := deaths[0].(amqp.Table)
		return death, ok
	case []amqp.Table:
		if len(deaths) == 0 {
			return nil, false
		}
		return deaths[0], true
	default:
		return nil, false
	}
}

func headerString(headers amqp.Table, key string) (string, bool) {
	value, exists := headers[key]
	if !exists {
		return "", false
	}
	text, ok := value.(string)
	return text, ok && text != ""
}

func headerInteger(headers amqp.Table, key string) (int, bool) {
	value, exists := headers[key]
	if !exists {
		return 0, false
	}
	switch typed := value.(type) {
	case int8:
		return int(typed), true
	case int16:
		return int(typed), true
	case int32:
		return int(typed), true
	case int64:
		if typed < 0 || uint64(typed) > uint64(^uint(0)>>1) {
			return 0, false
		}
		return int(typed), true
	case int:
		return typed, true
	default:
		return 0, false
	}
}
