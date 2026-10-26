package profile

import "testing"

func TestStorageEnumsAreClosed(t *testing.T) {
	for name, valid := range map[string]bool{
		"resource":       ResourceKind("future").Valid(),
		"logical type":   LogicalType("future").Valid(),
		"cardinality":    Cardinality("future").Valid(),
		"classification": Classification("future").Valid(),
		"normalization":  NormalizationProfile("future").Valid(),
		"field state":    FieldState("future").Valid(),
		"schema state":   SchemaState("future").Valid(),
	} {
		if valid {
			t.Fatalf("unknown %s was accepted", name)
		}
	}
}
