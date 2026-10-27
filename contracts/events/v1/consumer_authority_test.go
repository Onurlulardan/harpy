package eventsv1

import "testing"

func TestConsumerReceiptsRejectV1WriterAuthority(t *testing.T) {
	consumers := testConsumerCatalog(t, EffectExternal)
	producers := testProducerCatalog(t)
	tenant := eventTenantID
	v1Authority := consumerAuthority()
	v1Authority.WriterOrigin = WriterV1

	bootstrap := BootstrapReceipt{
		ReceiptID: receiptID, Consumer: "workflow.projector", ConsumerVersion: 1,
		TenantID: &tenant, ProducerAuthority: producerAuthority(), ConsumerAuthority: v1Authority,
		AggregateType: "form_submission", AggregateID: aggregateID, ThroughSequence: 7,
		StateReference: "snapshot/submission/42", StateSHA256: digest("b"),
		CorrelationID: "bootstrap:42", CausationID: causationID, OccurredAt: eventTime,
	}
	if _, err := NewBootstrapReceipt(bootstrap, consumers); err == nil {
		t.Fatal("bootstrap accepted a v1 consumer writer")
	}

	envelope, err := NewEnvelope(envelopeInput(), producers)
	if err != nil {
		t.Fatal(err)
	}
	effect := effectReceiptInput()
	effect.ConsumerAuthority = v1Authority
	effect.ExternalEffectReceiptID = eventString(effectID)
	if _, err := NewEffectReceipt(effect, envelope, producers, consumers); err == nil {
		t.Fatal("effect receipt accepted a v1 consumer writer")
	}

	operator := OperatorReceipt{
		ReceiptID: receiptID, Action: OperatorRetry, Consumer: "workflow.projector", ConsumerVersion: 1,
		TenantID: &tenant, ProducerAuthority: producerAuthority(), ConsumerAuthority: v1Authority,
		SourceEventID: eventID, ResultEventID: eventID,
		SourceBusinessEffectKey: "submission/42", ResultBusinessEffectKey: "submission/42",
		Reason: "verified", TicketReference: "OPS-42", ActorID: actorID,
		CausationID: causationID, OccurredAt: eventTime,
	}
	if _, err := NewOperatorReceipt(operator, consumers); err == nil {
		t.Fatal("operator receipt accepted a v1 consumer writer")
	}
}
