package systemresource

import (
	"encoding/json"
	"testing"
)

func TestFieldValuesAreCanonicalAndOwned(t *testing.T) {
	fieldID := mustSystemResourceID(t, "a1000000-0000-4000-8000-000000000001")
	input := map[ID]json.RawMessage{fieldID: []byte(`{"b":2,"a":1}`)}
	values, err := NewFieldValues(input)
	if err != nil {
		t.Fatal(err)
	}
	input[fieldID][2] = 'z'
	canonical, err := values.CanonicalJSON()
	if err != nil || string(canonical) !=
		`{"a1000000-0000-4000-8000-000000000001":{"a":1,"b":2}}` {
		t.Fatalf("canonical=%s error=%v", canonical, err)
	}
	owned, err := values.Values()
	if err != nil {
		t.Fatal(err)
	}
	owned[fieldID][0] = '['
	again, _ := values.CanonicalJSON()
	if string(again) != string(canonical) {
		t.Fatal("caller mutation changed canonical values")
	}
}

func TestFieldValuesRejectAmbiguousJSON(t *testing.T) {
	fieldID := mustSystemResourceID(t, "a1000000-0000-4000-8000-000000000001")
	if _, err := NewFieldValues(map[ID]json.RawMessage{
		fieldID: []byte(`{"a":1,"a":2}`),
	}); err == nil {
		t.Fatal("duplicate JSON member accepted")
	}
}
