package cutoverv1

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestChangeBatchCompletionBarrierAndPersistedOrdinals(t *testing.T) {
	barrier := validTypedChangeBatch()
	barrier.Payload.CheckpointState = "complete"
	barrier.Payload.Position.FinalLSN = &barrier.Payload.Position.CommitLSN
	barrier.Payload.Changes = []Change{}
	if err := ValidateChangeBatch(barrier); err != nil {
		t.Fatalf("valid completion barrier: %v", err)
	}
	assertChangeBatchSchema(t, barrier, true)

	nonterminal := barrier
	nonterminal.Payload.CheckpointState = "delta"
	nonterminal.Payload.Position.FinalLSN = nil
	if err := ValidateChangeBatch(nonterminal); err == nil {
		t.Fatal("nonterminal empty batch accepted")
	}
	assertChangeBatchSchema(t, nonterminal, false)

	differentFinal := "0/30"
	barrier.Payload.Position.FinalLSN = &differentFinal
	if err := ValidateChangeBatch(barrier); err == nil {
		t.Fatal("completion barrier before final LSN accepted")
	}

	tests := []struct {
		batch, local, want int
	}{
		{1, 1, 1},
		{1, MaximumChangeBatchChanges, 1000},
		{2, 1, 1001},
		{MaximumChangeBatchOrdinal, MaximumChangeBatchChanges, 2147483000},
	}
	for _, test := range tests {
		got, err := PersistedChangeOrdinal(test.batch, test.local)
		if err != nil || got != test.want {
			t.Fatalf("persisted ordinal (%d,%d) = %d, %v", test.batch, test.local, got, err)
		}
	}
	if _, err := PersistedChangeOrdinal(MaximumChangeBatchOrdinal+1, 1); err == nil {
		t.Fatal("overflowing batch ordinal accepted")
	}
}

func assertChangeBatchSchema(t *testing.T, batch ChangeBatch, valid bool) {
	t.Helper()
	document, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
	if err != nil {
		t.Fatal(err)
	}
	err = compileCutoverSchema(t, "change-batch.schema.json").Validate(decoded)
	if (err == nil) != valid {
		t.Fatalf("change batch schema validity = %t, want %t: %v", err == nil, valid, err)
	}
}
