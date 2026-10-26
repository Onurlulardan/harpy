package attestation

import (
	"encoding/json"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func TestVerifyCutoverChangeBatchAcceptsExplicitInsertAndDeleteNulls(t *testing.T) {
	_, _, lease, _ := verifiedCutoverLeaseFixture(t)
	tests := []struct {
		name   string
		kind   string
		before json.RawMessage
		after  json.RawMessage
	}{
		{name: "insert", kind: "insert", before: json.RawMessage("null"),
			after: json.RawMessage(`{"id":"1"}`)},
		{name: "delete", kind: "delete", before: json.RawMessage(`{"id":"1"}`),
			after: json.RawMessage("null")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := cutoverNullBatchDocument(t, test.kind, test.before, test.after)
			if _, err := VerifyCutoverChangeBatch(document, lease); err != nil {
				t.Fatalf("valid %s batch rejected: %v", test.kind, err)
			}
		})
	}
}

func TestVerifyCutoverChangeBatchRejectsWrongScalarAndNullShapes(t *testing.T) {
	_, _, lease, _ := verifiedCutoverLeaseFixture(t)
	tests := []struct {
		name   string
		kind   string
		before json.RawMessage
		after  json.RawMessage
	}{
		{name: "scalar before", kind: "insert", before: json.RawMessage("1"),
			after: json.RawMessage(`{"id":"1"}`)},
		{name: "insert null after", kind: "insert", before: json.RawMessage("null"),
			after: json.RawMessage("null")},
		{name: "update null before", kind: "update", before: json.RawMessage("null"),
			after: json.RawMessage(`{"id":"1"}`)},
		{name: "delete null before", kind: "delete", before: json.RawMessage("null"),
			after: json.RawMessage("null")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := cutoverNullBatchDocument(t, test.kind, test.before, test.after)
			if _, err := VerifyCutoverChangeBatch(document, lease); err == nil {
				t.Fatalf("invalid %s batch accepted", test.name)
			}
		})
	}
}

func TestVerifyCutoverChangeBatchRejectsUnplannedCheckpointKey(t *testing.T) {
	_, _, lease, _ := verifiedCutoverLeaseFixture(t)
	document := cutoverNullBatchDocument(
		t, "insert", json.RawMessage("null"), json.RawMessage(`{"id":"1"}`),
	)
	var batch cutoverv1.ChangeBatch
	if err := DecodeStrict(document, &batch); err != nil {
		t.Fatal(err)
	}
	batch.Payload.CheckpointKey = "unplanned"
	payload, err := MarshalCanonical(batch.Payload)
	if err != nil {
		t.Fatal(err)
	}
	batch.BatchSHA256, err = DigestHex(payload)
	if err != nil {
		t.Fatal(err)
	}
	document, err = MarshalCanonical(batch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyCutoverChangeBatch(document, lease); err == nil {
		t.Fatal("unplanned checkpoint key accepted")
	}
}

func cutoverNullBatchDocument(
	t *testing.T, kind string, before, after json.RawMessage,
) []byte {
	t.Helper()
	finalLSN := "0/20"
	payload := cutoverv1.ChangeBatchPayload{
		OperationID: attestationOperationID, RunID: attestationRunID,
		ScopeKind: "tenant", ScopeID: "66666666-6666-4666-8666-666666666666",
		RelationOrdinal: 1, CheckpointKey: "main", CheckpointState: "complete",
		Position: cutoverv1.SourcePosition{
			SnapshotName: "snapshot-main", SnapshotLSN: "0/10", ReceivedLSN: "0/20",
			CommitLSN: "0/20", FinalLSN: &finalLSN, SourceXID: "41", BatchOrdinal: 1,
		},
		LastKey: json.RawMessage(`{"id":"1"}`),
		Changes: []cutoverv1.Change{{Ordinal: 1, Kind: kind,
			Key: json.RawMessage(`{"id":"1"}`), Before: before, After: after}},
	}
	payloadDocument, err := MarshalCanonical(payload)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := DigestHex(payloadDocument)
	if err != nil {
		t.Fatal(err)
	}
	document, err := MarshalCanonical(cutoverv1.ChangeBatch{
		ContractVersion: "cutover-change-batch.v1", Payload: payload, BatchSHA256: digest,
	})
	if err != nil {
		t.Fatal(err)
	}
	return document
}
