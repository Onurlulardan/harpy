package profile

import (
	"bytes"
	"testing"
)

func TestNewFieldFromSnapshotVerifiesBothImmutableDigests(t *testing.T) {
	original := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	validation := original.ValidationSHA256()
	semantic := original.SemanticSHA256()
	input := FieldSnapshotInput{
		ID: original.ID(), Resource: original.Resource(), Key: original.Key(),
		Type: original.Type(), Cardinality: original.Cardinality(),
		Required: original.Required(), Classification: original.Classification(),
		Normalization: original.Normalization(), State: original.State(),
		ValidationJSON:           original.ValidationJSON(),
		ExpectedValidationSHA256: validation[:], ExpectedSemanticSHA256: semantic[:],
	}
	hydrated, err := NewFieldFromSnapshot(input)
	if err != nil || hydrated.SemanticSHA256() != semantic ||
		!bytes.Equal(hydrated.ValidationJSON(), original.ValidationJSON()) {
		t.Fatalf("hydrate exact snapshot: %v", err)
	}

	tests := map[string]func(*FieldSnapshotInput){
		"key drift":        func(value *FieldSnapshotInput) { value.Key = "changed_key" },
		"validation drift": func(value *FieldSnapshotInput) { value.ValidationJSON = []byte(`{"maximum_bytes":12}`) },
		"semantic drift":   func(value *FieldSnapshotInput) { value.ExpectedSemanticSHA256 = make([]byte, 32) },
		"unknown contract": func(value *FieldSnapshotInput) {
			value.ValidationJSON = []byte(`{"maximum_bytes":1048576,"unknown":1}`)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := input
			mutate(&candidate)
			if _, err := NewFieldFromSnapshot(candidate); err == nil {
				t.Fatal("drifted field snapshot was accepted")
			}
		})
	}
}
