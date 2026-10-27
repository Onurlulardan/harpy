package adoption

import "testing"

func TestEncodePlanProducesStableCanonicalEvidence(t *testing.T) {
	compiler, err := NewCompiler(testNormalizer, testNormalizer)
	if err != nil {
		t.Fatalf("create compiler: %v", err)
	}
	snapshot, mapping := validFixture()
	plan, err := compiler.Evaluate(snapshot, mapping)
	if err != nil {
		t.Fatalf("evaluate fixture: %v", err)
	}
	first, firstSHA, err := EncodePlan(plan)
	if err != nil {
		t.Fatalf("encode plan: %v", err)
	}
	second, secondSHA, err := EncodePlan(plan)
	if err != nil {
		t.Fatalf("encode plan again: %v", err)
	}
	if string(first) != string(second) || firstSHA != secondSHA || len(firstSHA) != 64 {
		t.Fatalf("plan document is not stable: %s %s", firstSHA, secondSHA)
	}
}
