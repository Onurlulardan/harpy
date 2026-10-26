package rowdecode

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestValueIsImmutableAndRedactedOnAmbientSurfaces(t *testing.T) {
	value, err := Decode(textType(t), "top-secret")
	if err != nil {
		t.Fatal(err)
	}
	for _, rendered := range []string{
		fmt.Sprint(value), fmt.Sprintf("%#v", value), fmt.Sprintf("%+v", value),
		value.LogValue().String(), slog.Any("value", value).Value.String(),
	} {
		if rendered != redacted || strings.Contains(rendered, "top-secret") {
			t.Fatalf("value leaked through formatting: %q", rendered)
		}
	}
	encoded, err := json.Marshal(value)
	if err == nil || len(encoded) != 0 || strings.Contains(err.Error(), "top-secret") {
		t.Fatalf("direct JSON marshal did not fail redacted: %q, %v", encoded, err)
	}

	response, _ := value.ResponseJSON()
	response[1] = 'X'
	requireResponse(t, value, `"top-secret"`)
}

func TestDecodeRowIsPositionalAtomicAndReturnsOwnedValues(t *testing.T) {
	booleanType, _ := api.NewBooleanFieldType()
	types := []api.FieldType{textType(t), booleanType}
	values, err := DecodeRow(types, []any{"owned", true})
	if err != nil || len(values) != 2 {
		t.Fatalf("decode row: %v", err)
	}
	requireResponse(t, values[0], `"owned"`)
	requireResponse(t, values[1], "true")
	if result, err := DecodeRow(types, []any{"owned", "true"}); err == nil || result != nil {
		t.Fatal("partial row returned")
	}
	for _, invalid := range []struct {
		types  []api.FieldType
		values []any
	}{
		{nil, nil}, {types, []any{"only-one"}}, {types[:1], []any{"too", true}},
	} {
		if result, err := DecodeRow(invalid.types, invalid.values); err == nil || result != nil {
			t.Fatal("invalid row accepted")
		}
	}
}
