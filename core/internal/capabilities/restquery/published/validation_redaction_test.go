package published

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestPageOverrideBoundariesAndValidationAwareAccessors(t *testing.T) {
	for _, size := range []int{1, 500} {
		page, err := NewPageOverride(&size, nil)
		if err != nil {
			t.Fatalf("valid size %d rejected: %v", size, err)
		}
		value, present, err := page.Size()
		if err != nil || !present || value != size {
			t.Fatalf("size = %d, present = %v, err = %v", value, present, err)
		}
	}
	validCursor := strings.Repeat("x", maximumPublishedCursorBytes)
	if _, err := NewPageOverride(nil, &validCursor); err != nil {
		t.Fatalf("cursor boundary rejected: %v", err)
	}
	invalidCursors := []string{"", "contains\x00nul", validCursor + "x"}
	for _, cursor := range invalidCursors {
		if _, err := NewPageOverride(nil, &cursor); err == nil {
			t.Fatal("invalid cursor accepted")
		}
	}
	for _, size := range []int{0, 501} {
		if _, err := NewPageOverride(&size, nil); err == nil {
			t.Fatalf("invalid size %d accepted", size)
		}
	}
	if _, err := NewPageOverride(nil, nil); err == nil {
		t.Fatal("empty page accepted")
	}
	var zero PageOverride
	if _, _, err := zero.Size(); err == nil {
		t.Fatal("zero page exposed size")
	}
	if _, _, err := zero.After(); err == nil {
		t.Fatal("zero page exposed cursor")
	}
}

func TestBoundAndDefinitionAccessorsFailClosed(t *testing.T) {
	var bound BoundParameters
	if _, _, err := bound.Lookup("secret"); err == nil {
		t.Fatal("zero bound set allowed lookup")
	}
	if _, err := bound.Names(); err == nil {
		t.Fatal("zero bound set exposed names")
	}
	if _, _, err := bound.Page(); err == nil {
		t.Fatal("zero bound set exposed page")
	}
	var definition ParameterDefinition
	if _, err := definition.Name(); err == nil {
		t.Fatal("zero definition exposed name")
	}
	if _, err := definition.Type(); err == nil {
		t.Fatal("zero definition exposed type")
	}
	if _, err := definition.Arity(); err == nil {
		t.Fatal("zero definition exposed arity")
	}
	if _, err := definition.Required(); err == nil {
		t.Fatal("zero definition exposed required state")
	}
	if _, _, err := definition.Default(); err == nil {
		t.Fatal("zero definition exposed default")
	}
}

func TestPublishedValuesRedactFormattingAndRejectDirectJSON(t *testing.T) {
	manifest := testBindingManifest(t)
	bound, err := BindGET(manifest, bindingValues(t, "7"), nil)
	if err != nil {
		t.Fatal(err)
	}
	definition, found, err := manifest.Lookup("label")
	if err != nil || !found {
		t.Fatal("definition missing")
	}
	secret := "do-not-leak"
	page, err := NewPageOverride(nil, &secret)
	if err != nil {
		t.Fatal(err)
	}
	values := []any{
		ParameterInput{Name: secret, Type: testTextFieldType(t),
			Arity: ParameterAritySingle},
		definition, manifest, page, bound,
	}
	for _, value := range values {
		formatted := fmt.Sprintf("%v|%+v|%#v|%s", value, value, value, value)
		if strings.Contains(formatted, secret) ||
			formatted != "[REDACTED]|[REDACTED]|[REDACTED]|[REDACTED]" {
			t.Fatalf("format leaked %T: %q", value, formatted)
		}
		if encoded, err := json.Marshal(value); err == nil || encoded != nil {
			t.Fatalf("JSON marshaled %T: %q, %v", value, encoded, err)
		}
	}
	if encoded, err := json.Marshal(ParameterAritySingle); err == nil || encoded != nil {
		t.Fatalf("arity JSON = %q, %v", encoded, err)
	}
}

func TestInvalidResolvedValueCannotEnterBoundParameters(t *testing.T) {
	definition := ParameterDefinition{
		name: "value", fieldType: testTextFieldType(t),
		arity: ParameterAritySingle, validated: true,
	}
	bound := BoundParameters{
		parameters: []boundParameter{{
			definition: definition,
			values:     []resolved.Value{mustScalarValue(t, api.LogicalTypeInteger, "1")},
		}},
		validated: true,
	}
	if err := bound.Validate(); err == nil {
		t.Fatal("mismatched resolved value validated")
	}
}
