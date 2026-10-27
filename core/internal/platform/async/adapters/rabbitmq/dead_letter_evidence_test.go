package rabbitmq

import (
	"testing"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDeadLetterEvidenceClassifiesBoundedBrokerTerminalReasons(t *testing.T) {
	tests := []struct {
		reason string
		class  deadletter.ReasonClass
	}{
		{reason: "rejected", class: deadletter.ReasonPermanent},
		{reason: "delivery_limit", class: deadletter.ReasonDeliveryLimit},
		{reason: "expired", class: deadletter.ReasonDeliveryLimit},
	}
	for _, test := range tests {
		t.Run(test.reason, func(t *testing.T) {
			evidence, err := deadLetterEvidence(amqp.Table{
				"x-death": []any{amqp.Table{
					"reason": test.reason, "count": int64(7),
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if evidence.Reason != test.class || evidence.DeliveryCount != 7 {
				t.Fatalf("evidence = %#v", evidence)
			}
		})
	}
}

func TestDeadLetterEvidenceRejectsUnknownOrMalformedHeaders(t *testing.T) {
	for _, headers := range []amqp.Table{
		{},
		{"x-death": "not-an-array"},
		{"x-death": []any{amqp.Table{"reason": "rejected", "count": "many"}}},
		{"x-first-death-reason": "operator_deleted"},
	} {
		if _, err := deadLetterEvidence(headers); err == nil {
			t.Fatalf("invalid evidence accepted: %#v", headers)
		}
	}
}
