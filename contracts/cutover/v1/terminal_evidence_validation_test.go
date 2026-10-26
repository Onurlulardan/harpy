package cutoverv1

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestValidateTerminalEvidenceAcceptsExactCompletionProof(t *testing.T) {
	if err := ValidateTerminalEvidence(validTypedTerminalEvidence()); err != nil {
		t.Fatalf("valid terminal evidence rejected: %v", err)
	}
}

func TestValidateTerminalEvidenceRejectsSemanticDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*TerminalEvidence)
	}{
		{"lease manifest mismatch", func(v *TerminalEvidence) {
			v.InitialCopy.ManifestSHA256 = strings.Repeat("b", 64)
		}},
		{"non canonical final lsn", func(v *TerminalEvidence) { *v.FinalLSN = "0/020" }},
		{"initial copy after final", func(v *TerminalEvidence) { v.InitialCopy.SnapshotLSN = "0/30" }},
		{"partition underflow", func(v *TerminalEvidence) { v.InitialCopy.PartitionCount = 0 }},
		{"checkpoint count mismatch", func(v *TerminalEvidence) { *v.CheckpointCount++ }},
		{"checkpoint before final", func(v *TerminalEvidence) { v.Checkpoints[0].AppliedLSN = "0/10" }},
		{"checkpoint key control", func(v *TerminalEvidence) { v.Checkpoints[0].CheckpointKey = "main\n" }},
		{"journal without changes", func(v *TerminalEvidence) { *v.AppliedChangeCount = 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := validTypedTerminalEvidence()
			test.mutate(&value)
			if err := ValidateTerminalEvidence(value); err == nil {
				t.Fatal("invalid terminal evidence accepted")
			}
		})
	}
}

func TestValidateTerminalEvidenceAcceptsExactAbortProofs(t *testing.T) {
	digest := strings.Repeat("a", 64)
	tests := []struct{ state, code string }{
		{"failed", "verification_failed"},
		{"cancelled", "operator_cancelled"},
	}
	for _, test := range tests {
		value := TerminalEvidence{
			ContractVersion: "cutover-terminal-evidence.v1",
			OperationID:     typedOperationID, RunID: typedRunID,
			LeaseManifestSHA256: digest, PlanSHA256: digest,
			TransformManifestSHA256: digest, TerminalState: test.state,
			OutcomeCode: &test.code, OutcomeEvidenceSHA256: &digest,
		}
		if err := ValidateTerminalEvidence(value); err != nil {
			t.Fatalf("valid %s terminal evidence rejected: %v", test.state, err)
		}
		finalLSN := "0/20"
		value.FinalLSN = &finalLSN
		if err := ValidateTerminalEvidence(value); err == nil {
			t.Fatalf("mixed %s completion evidence accepted", test.state)
		}
	}
}

func TestTerminalEvidenceSchemaPinsCompletedFourteenAndAbortNineKeys(t *testing.T) {
	completed := validTypedTerminalEvidence()
	assertTerminalSchemaShape(t, completed, 14, true)
	digest, code := strings.Repeat("a", 64), "verification_failed"
	abort := TerminalEvidence{
		ContractVersion: "cutover-terminal-evidence.v1",
		OperationID:     typedOperationID, RunID: typedRunID,
		LeaseManifestSHA256: digest, PlanSHA256: digest,
		TransformManifestSHA256: digest, TerminalState: "failed",
		OutcomeCode: &code, OutcomeEvidenceSHA256: &digest,
	}
	assertTerminalSchemaShape(t, abort, 9, true)
	abort.FinalLSN = completed.FinalLSN
	assertTerminalSchemaShape(t, abort, 10, false)
}

func assertTerminalSchemaShape(
	t *testing.T, value TerminalEvidence, expectedKeys int, accepted bool,
) {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(document, &members); err != nil {
		t.Fatal(err)
	}
	if len(members) != expectedKeys {
		t.Fatalf("terminal evidence key count = %d, want %d", len(members), expectedKeys)
	}
	decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
	if err != nil {
		t.Fatal(err)
	}
	err = compileCutoverSchema(t, "terminal-evidence.schema.json").Validate(decoded)
	if accepted && err != nil {
		t.Fatalf("valid terminal schema rejected: %v", err)
	}
	if !accepted && err == nil {
		t.Fatal("mixed terminal schema accepted")
	}
}

func TestTerminalRootFramingIsStableAndUnambiguous(t *testing.T) {
	digest := strings.Repeat("a", 64)
	batchRoot, err := TerminalBatchRootSHA256([]TerminalBatchLeaf{{
		BatchOrdinal: 1, SourceXID: "42", SourceCommitLSN: "0/20",
		BatchSHA256: digest, AppliedCount: 1,
		KeyStateSHA256: digest, StateSHA256: digest,
	}})
	if err != nil {
		t.Fatal(err)
	}
	coverageRoot, count, err := TerminalJournalCoverageRootSHA256(
		[]TerminalJournalCoverageLeaf{{
			SourceXID: "42", SourceCommitLSN: "0/20",
			Changes: []TerminalChangeLeaf{{
				RelationOrdinal: 1, ChangeOrdinal: 1, CheckpointKey: "a:1",
				MutationKind: "insert", MutationSHA256: digest,
			}},
		}},
	)
	if err != nil || count != 1 {
		t.Fatalf("journal root failed: count=%d err=%v", count, err)
	}
	if batchRoot != "51634dd55cec5c46be031b30032d8e6addeef630599828f511130e619c47b12c" ||
		coverageRoot != "c0c7c2bf1d50421a4a1d5d4bde423daf4d3a4963ba41b5a04fd2e11dc73cf09c" {
		t.Fatalf("terminal root framing drift: batch=%s coverage=%s", batchRoot, coverageRoot)
	}
}

func TestTerminalRootRejectsReorderedLeaves(t *testing.T) {
	digest := strings.Repeat("a", 64)
	batches := []TerminalBatchLeaf{
		{BatchOrdinal: 2, SourceXID: "42", SourceCommitLSN: "0/20",
			BatchSHA256: digest, KeyStateSHA256: digest, StateSHA256: digest},
		{BatchOrdinal: 1, SourceXID: "43", SourceCommitLSN: "0/30",
			BatchSHA256: digest, KeyStateSHA256: digest, StateSHA256: digest},
	}
	if _, err := TerminalBatchRootSHA256(batches); err == nil {
		t.Fatal("reordered batch leaf accepted")
	}
	leaves := []TerminalJournalCoverageLeaf{
		{SourceXID: "43", SourceCommitLSN: "0/20", Changes: []TerminalChangeLeaf{{
			RelationOrdinal: 1, ChangeOrdinal: 1, CheckpointKey: "main",
			MutationKind: "insert", MutationSHA256: digest,
		}}},
		{SourceXID: "42", SourceCommitLSN: "0/30", Changes: []TerminalChangeLeaf{{
			RelationOrdinal: 1, ChangeOrdinal: 2, CheckpointKey: "main",
			MutationKind: "update", MutationSHA256: digest,
		}}},
	}
	if _, _, err := TerminalJournalCoverageRootSHA256(leaves); err == nil {
		t.Fatal("reordered journal leaves accepted")
	}
}
