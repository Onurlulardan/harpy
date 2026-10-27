package eventsv1

import "testing"

func TestBrokerTransportIsVersionedAndDeterministic(t *testing.T) {
	key, err := RoutingKey("forms.entry.created", 3)
	if err != nil {
		t.Fatal(err)
	}
	if key != "forms.entry.created.v3" || BrokerExchangeName != "harpy.events.v1" ||
		DeadLetterExchangeName != "harpy.dead_letters.v1" {
		t.Fatal("broker transport contract drifted")
	}
	if _, err := RoutingKey("INVALID", 1); err == nil {
		t.Fatal("invalid event type produced a routing key")
	}
}
