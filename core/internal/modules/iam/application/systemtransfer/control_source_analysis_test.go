package systemtransfer

import "testing"

func TestSourceAnalysisOwnedCopyIsIsolated(t *testing.T) {
	value := SourceAnalysis{Columns: []SourceColumn{{Name: "email", Kind: "string"}},
		Samples: []map[string]any{{"email": "one@example.test"}}}
	if err := value.Validate(); err != nil {
		t.Fatal(err)
	}
	owned := value.OwnedCopy()
	value.Columns[0].Name = "changed"
	value.Samples[0]["email"] = "changed@example.test"
	if owned.Columns[0].Name != "email" ||
		owned.Samples[0]["email"] != "one@example.test" {
		t.Fatalf("copy shares source memory: %#v", owned)
	}
}

func TestSourceAnalysisOwnedCopyFailsClosedForUnvalidatedValue(t *testing.T) {
	value := SourceAnalysis{Columns: []SourceColumn{{Name: "bad", Kind: "string"}},
		Samples: []map[string]any{{"bad": make(chan int)}}}
	owned := value.OwnedCopy()
	if len(owned.Columns) != 0 || len(owned.Samples) != 0 {
		t.Fatalf("invalid copy was exposed: %#v", owned)
	}
}
