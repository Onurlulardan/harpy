package rowdecode

import (
	"encoding/json"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func TestDecodeListsPreservesOrderDuplicatesAndEmpty(t *testing.T) {
	text, err := Decode(textListType(t), []any{"a", "a", "Caf\u00e9"})
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, text, "[\"a\",\"a\",\"Caf\u00e9\"]")
	empty, err := Decode(textListType(t), []any{})
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, empty, `[]`)

	first := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	second := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
	fieldType, _ := api.NewUUIDListFieldType()
	identifiers, err := Decode(fieldType, []any{[16]byte(first), [16]byte(second)})
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, identifiers,
		`["123e4567-e89b-12d3-a456-426614174000","123e4567-e89b-12d3-a456-426614174001"]`)
}

func TestDecodeListsRejectsElementAndContainerShapeDrift(t *testing.T) {
	textCases := []any{
		[]string{"a"}, []any{[]byte("a")}, []any{nil},
		[]any{"e\u0301"}, []any{"a\x00b"},
	}
	for _, candidate := range textCases {
		if _, err := Decode(textListType(t), candidate); err == nil {
			t.Fatal("invalid text list accepted")
		}
	}
	uuidType, _ := api.NewUUIDListFieldType()
	identifier := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	for _, candidate := range []any{[]uuid.UUID{identifier}, []any{identifier}, []any{nil}, []any{[16]byte{}}} {
		if _, err := Decode(uuidType, candidate); err == nil {
			t.Fatal("invalid UUID list accepted")
		}
	}
}

func TestDecodeJSONRequiresRawBytesAndCanonicalizesRFC8785(t *testing.T) {
	fieldType, _ := api.NewJSONFieldType()
	source := []byte(`{"b":2,"a":[true,"ok"]}`)
	value, err := Decode(fieldType, source)
	if err != nil {
		t.Fatal(err)
	}
	source[2] = 'x'
	requireResponse(t, value, `{"a":[true,"ok"],"b":2}`)
	response, _ := value.ResponseJSON()
	response[2] = 'x'
	requireResponse(t, value, `{"a":[true,"ok"],"b":2}`)

	raw := json.RawMessage(`{"z":0,"a":1}`)
	canonical, err := Decode(fieldType, raw)
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, canonical, `{"a":1,"z":0}`)
}

func TestDecodeJSONRejectsLossyOrMalformedShapes(t *testing.T) {
	fieldType, _ := api.NewJSONFieldType()
	cases := []any{
		map[string]any{"large": float64(9007199254740993)},
		`{"a":1}`,
		[]byte(`{"a":`),
		[]byte(`{"a":1,"a":2}`),
		[]byte(`{"a":"\u0000"}`),
		[]byte(`NaN`),
	}
	for _, candidate := range cases {
		if _, err := Decode(fieldType, candidate); err == nil {
			t.Fatalf("invalid JSON %T accepted", candidate)
		}
	}
	jsonNull, err := Decode(fieldType, []byte(`null`))
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, jsonNull, "null")
	isSQLNull, err := jsonNull.IsNull()
	if err != nil || isSQLNull {
		t.Fatalf("JSON literal null became SQL NULL: %v, %v", isSQLNull, err)
	}

	nullValue, err := Decode(fieldType, nil)
	if err != nil {
		t.Fatal(err)
	}
	requireResponse(t, nullValue, "null")
	isSQLNull, err = nullValue.IsNull()
	if err != nil || !isSQLNull {
		t.Fatalf("SQL NULL identity lost: %v, %v", isSQLNull, err)
	}
	if _, present, err := nullValue.LastValue(); err != nil || present {
		t.Fatalf("JSON unexpectedly cursor-capable: %v, %v", present, err)
	}
}
