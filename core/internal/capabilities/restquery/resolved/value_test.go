package resolved

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolvedValueKindsAndCopies(t *testing.T) {
	empty, err := NewScalarValue(api.LogicalTypeText, "")
	if err != nil {
		t.Fatalf("empty text must remain a real value: %v", err)
	}
	if scalar, found, err := empty.Scalar(); err != nil || !found || scalar != "" {
		t.Fatal("empty text scalar unavailable")
	}
	integer, err := NewScalarValue(api.LogicalTypeInteger, "42")
	if err != nil {
		t.Fatal(err)
	}
	if scalar, found, err := integer.Scalar(); err != nil || !found || scalar != "42" {
		t.Fatal("integer scalar unavailable")
	}
	boolean, err := NewBooleanValue(true)
	if err != nil {
		t.Fatal(err)
	}
	if value, found, err := boolean.Boolean(); err != nil || !found || !value {
		t.Fatal("boolean unavailable")
	}
	jsonValue, err := NewJSONValue([]byte(`{"b":2,"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	document, found, err := jsonValue.JSON()
	if err != nil || !found || string(document) != `{"a":1,"b":2}` {
		t.Fatalf("canonical JSON = %q", document)
	}
	document[0] = '['
	second, _, _ := jsonValue.JSON()
	if string(second) != `{"a":1,"b":2}` {
		t.Fatal("JSON accessor leaked mutable storage")
	}
}

func TestResolvedValueRejectsNonCanonicalScalars(t *testing.T) {
	cases := []struct {
		kind  api.LogicalType
		value string
	}{
		{api.LogicalTypeInteger, "01"}, {api.LogicalTypeInteger, "-0"},
		{api.LogicalTypeDecimal, "1e2"}, {api.LogicalTypeDecimal, "-0.00"},
		{api.LogicalTypeDate, "2026-02-30"},
		{api.LogicalTypeTimestamp, "2026-01-01T00:00:00Z"},
		{api.LogicalTypeUUID, "550E8400-E29B-41D4-A716-446655440000"},
		{api.LogicalTypeUUID, "00000000-0000-0000-0000-000000000001"},
		{api.LogicalTypeBoolean, "true"}, {api.LogicalTypeTextList, "value"},
	}
	for _, test := range cases {
		if _, err := NewScalarValue(test.kind, test.value); err == nil {
			t.Fatalf("accepted %s %q", test.kind, test.value)
		}
	}
}

func TestResolvedJSONRejectsNULAndInvalidNumbers(t *testing.T) {
	for _, document := range [][]byte{
		[]byte(`{"value":"\u0000"}`), []byte(`{"value":1e1000}`),
		[]byte(`{"x":1,"x":2}`), []byte(`null`), []byte(`null trailing`),
	} {
		if _, err := NewJSONValue(document); err == nil {
			t.Fatalf("accepted invalid JSON %q", document)
		}
	}
}

func TestResolvedValueFailClosedAndRedacted(t *testing.T) {
	invalid := []Value{{}, {kind: api.LogicalTypeInteger, state: valueStateScalar, scalar: "1"}}
	for _, value := range invalid {
		if value.Validate() == nil {
			t.Fatal("invalid value validated")
		}
		if _, _, err := value.Scalar(); err == nil {
			t.Fatal("invalid accessor succeeded")
		}
	}
	value, _ := NewScalarValue(api.LogicalTypeText, "secret")
	if fmt.Sprint(value) != valueRedaction || fmt.Sprintf("%#v", value) != valueRedaction {
		t.Fatal("resolved value leaked through formatting")
	}
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("resolved value silently marshaled")
	}
}
