package v1query

import "testing"

func TestResultRetirementKeysAreUniqueCanonicalAndOwned(t *testing.T) {
	result, err := mustDecoder(t).Decode(
		"q=term&sort[field]=legacy_status&" +
			"filters[legacy_status][operator]=same_as&" +
			"filters[legacy_status][value]=OPEN",
	)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := result.RetirementKeys()
	if err != nil || len(keys) != 4 {
		t.Fatalf("keys=%d err=%v", len(keys), err)
	}
	previous := ""
	for _, key := range keys {
		value, valueErr := key.Value()
		if valueErr != nil || previous >= value {
			t.Fatalf("non-canonical retirement keys: previous=%q value=%q", previous, value)
		}
		previous = value
	}
	keys[0] = RetirementKey{}
	again, err := result.RetirementKeys()
	if err != nil || again[0].Validate() != nil {
		t.Fatal("caller mutation changed stored retirement identity")
	}
}

func TestResultWithoutAliasesHasEmptyRetirementBatch(t *testing.T) {
	result, err := mustDecoder(t).Decode("search=term")
	if err != nil {
		t.Fatal(err)
	}
	keys, err := result.RetirementKeys()
	if err != nil || keys == nil || len(keys) != 0 {
		t.Fatalf("keys=%#v err=%v", keys, err)
	}
}
