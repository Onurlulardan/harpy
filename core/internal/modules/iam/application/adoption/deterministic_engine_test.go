package adoption

import "testing"

func TestCompilerBuildsDeterministicEligiblePlan(t *testing.T) {
	compiler, err := NewCompiler(testNormalizer, testNormalizer)
	if err != nil {
		t.Fatalf("create compiler: %v", err)
	}
	snapshot, mapping := validFixture()
	first, err := compiler.Evaluate(snapshot, mapping)
	if err != nil {
		t.Fatalf("compile valid fixture: %v", err)
	}
	if !first.Eligible || len(first.Issues) != 0 || len(first.Categories) != 15 {
		t.Fatalf("valid fixture was fenced: %+v", first.Issues)
	}
	reverseSnapshots(&snapshot)
	reverseMappings(&mapping)
	second, err := compiler.Evaluate(snapshot, mapping)
	if err != nil {
		t.Fatalf("compile reordered fixture: %v", err)
	}
	if first.SourceSHA256 != second.SourceSHA256 || first.MappingSHA256 != second.MappingSHA256 {
		t.Fatalf("canonical roots changed after reordering: first=%+v second=%+v", first, second)
	}
}

func reverseSnapshots(snapshot *Snapshot) {
	reverse(snapshot.Resources)
	reverse(snapshot.Actions)
	reverse(snapshot.Users[0].RoleIDs)
}

func reverseMappings(mapping *Mapping) {
	reverse(mapping.Resources)
	reverse(mapping.Actions)
	reverse(mapping.Identity.Users[0].Assignments)
}

func reverse[T any](values []T) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
