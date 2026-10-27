package authorization

import "testing"

func TestRequirementKindsAreExhaustive(t *testing.T) {
	expected := []string{
		"resource_read", "field_select", "field_filter", "field_sort",
		"field_search", "field_group", "field_measure", "field_having",
		"relation_include", "exact_count", "aggregate_measure",
		"aggregate_having", "aggregate_sort",
	}
	for index, value := range expected {
		kind := RequirementKind(index + 1)
		if err := kind.Validate(); err != nil || kind.String() != value {
			t.Fatalf("kind %d = %q, %v", index+1, kind.String(), err)
		}
	}
	if RequirementKind(0).Validate() == nil || RequirementKind(255).Validate() == nil {
		t.Fatal("unknown requirement kind was accepted")
	}
}
