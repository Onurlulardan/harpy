package jobsv1

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestStaleJobDispositionExactDecisionMatrix(t *testing.T) {
	adopt, err := NewStaleJobDisposition(staleDispositionFixture(StaleJobAdopt))
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(adopt)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeStaleJobDisposition(document)
	if err != nil || decoded.Decision != StaleJobAdopt {
		t.Fatalf("decode adopt: value=%+v err=%v", decoded, err)
	}

	cancel := staleDispositionFixture(StaleJobCancel)
	if _, err := NewStaleJobDisposition(cancel); err != nil {
		t.Fatal(err)
	}
	cancel.ResultJobID = stringPointer(retryJobID)
	if _, err := NewStaleJobDisposition(cancel); err == nil {
		t.Fatal("cancel accepted a result job")
	}

	badAuthority := staleDispositionFixture(StaleJobAdopt)
	badAuthority.CurrentAuthority = badAuthority.OldAuthority
	if _, err := NewStaleJobDisposition(badAuthority); err == nil {
		t.Fatal("non-stale authority transition was accepted")
	}
	databaseWithKey := staleDispositionFixture(StaleJobAdopt)
	databaseWithKey.EffectClass = EffectDatabaseTransactional
	if _, err := NewStaleJobDisposition(databaseWithKey); err == nil {
		t.Fatal("database job accepted an external effect key")
	}
	externalWithoutKey := staleDispositionFixture(StaleJobAdopt)
	externalWithoutKey.SourcePayload.EffectKey = nil
	externalWithoutKey.ResultPayload.EffectKey = nil
	if _, err := NewStaleJobDisposition(externalWithoutKey); err == nil {
		t.Fatal("external job accepted a missing effect key")
	}

	unknown := bytes.Replace(document, []byte(`"decision":"adopt"`),
		[]byte(`"decision":"retry"`), 1)
	if _, err := DecodeStaleJobDisposition(unknown); err == nil {
		t.Fatal("unknown disposition was accepted")
	}
}

func staleDispositionFixture(decision StaleJobDecision) StaleJobDisposition {
	effectKey := "form.export/tenant/111"
	oldAuthority := testAuthority()
	oldAuthority.GroupGeneration--
	oldAuthority.CutoverGeneration--
	source := StaleJobPayloadIdentity{
		Revision: 1, SHA256: hexDigest("a"), Reference: "payload/jobs/222",
		DedupeKey: "form.export/111", EffectKey: &effectKey,
	}
	value := StaleJobDisposition{
		EvidenceRootID: receiptID, DispositionID: attemptID,
		SourceJobID: jobID, Decision: decision,
		Kind: "form.export", EffectClass: EffectExternal,
		MaximumAttempts: 8, Priority: 10,
		OldAuthority: oldAuthority, CurrentAuthority: testAuthority(),
		SourcePayload: source, Reason: "operator verified durable evidence",
		SourceRunAt:     timestamp,
		TicketReference: "OPS-42", ActorID: actorID,
		CausationID: causationID, OccurredAt: timestamp,
	}
	if decision == StaleJobAdopt {
		result := source
		result.Reference = "payload/jobs/333"
		result.DedupeKey = "form.export/111:g4:d3"
		value.ResultJobID = stringPointer(retryJobID)
		value.ResultPayload = &result
		value.ResultRunAt = stringPointer(timestamp)
	}
	return value
}
