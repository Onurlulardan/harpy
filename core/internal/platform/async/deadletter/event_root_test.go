package deadletter

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestEventRootIsDeterministicAndKeepsPayloadOutsideRedactedEnvelope(t *testing.T) {
	document, consumers := eventRootFixture(t, "forms.entry.created")
	evidence := DeliveryEvidence{Reason: ReasonPermanent, DeliveryCount: 2}
	first, err := NewEventRoot(document, "workflow.projector.v1", evidence, consumers)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewEventRoot(document, "workflow.projector.v1", evidence, consumers)
	if err != nil {
		t.Fatal(err)
	}
	if first.RootID() != second.RootID() || first.ReasonCode() != EventPermanentCode {
		t.Fatalf("dead-letter identity drifted: %#v %#v", first, second)
	}
	var metadata map[string]any
	if err := json.Unmarshal(first.RedactedEnvelope(), &metadata); err != nil {
		t.Fatal(err)
	}
	if _, exists := metadata["payload"]; exists || !bytes.Equal(
		first.Payload(), []byte(`{"record_id":"42"}`),
	) {
		t.Fatal("dead-letter payload leaked into redacted envelope or drifted")
	}
}

func TestEventRootClassifiesUnsupportedSubscriptionAsUnknownVersion(t *testing.T) {
	document, consumers := eventRootFixture(t, "forms.entry.deleted")
	root, err := NewEventRoot(
		document, "workflow.projector.v1",
		DeliveryEvidence{Reason: ReasonPermanent, DeliveryCount: 1}, consumers,
	)
	if err != nil {
		t.Fatal(err)
	}
	if root.ReasonClass() != ReasonUnknownVersion || root.ReasonCode() != EventVersionCode {
		t.Fatalf("unknown subscription classified as %s/%s", root.ReasonClass(), root.ReasonCode())
	}
}

func TestEventRootRejectsUnregisteredQueueAndReason(t *testing.T) {
	document, consumers := eventRootFixture(t, "forms.entry.created")
	if _, err := NewEventRoot(
		document, "missing.queue.v1",
		DeliveryEvidence{Reason: ReasonPermanent}, consumers,
	); err == nil {
		t.Fatal("unregistered queue was accepted")
	}
	if _, err := NewEventRoot(
		document, "workflow.projector.v1",
		DeliveryEvidence{Reason: ReasonUnknownVersion}, consumers,
	); err == nil {
		t.Fatal("caller-supplied unknown-version reason was accepted")
	}
}
