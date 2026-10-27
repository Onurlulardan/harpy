package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestProjectSystemResourceQueryItemTranslatesAllResourceRevisionWires(t *testing.T) {
	tests := []struct {
		resource string
		field    string
		value    string
	}{
		{"users", "email", "superadmin@harpy.io"},
		{"roles", "name", "super-admin"},
		{"organizations", "slug", "harpy"},
	}
	for _, test := range tests {
		t.Run(test.resource, func(t *testing.T) {
			allowed := map[string]struct{}{
				"id": {}, "revision": {}, test.field: {},
			}
			input := json.RawMessage(fmt.Sprintf(
				`{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"7","%s":%q}`,
				test.field, test.value,
			))
			original := bytes.Clone(input)
			projected, valid := projectSystemResourceQueryItem(input, allowed)
			if !valid {
				t.Fatal("canonical REST Query revision was rejected")
			}
			var response map[string]any
			if err := json.Unmarshal(projected, &response); err != nil {
				t.Fatal(err)
			}
			if response["revision"] != float64(7) || response[test.field] != test.value {
				t.Fatalf("projected response = %v", response)
			}
			if strings.Contains(string(projected), `"revision":"7"`) ||
				!strings.Contains(string(projected), `"revision":7`) {
				t.Fatalf("revision was not projected to the HTTP integer contract: %s", projected)
			}
			if !bytes.Equal(input, original) ||
				!strings.Contains(string(input), `"revision":"7"`) {
				t.Fatalf("REST Query wire was mutated: %s", input)
			}
		})
	}
}

func TestProjectSystemResourceQueryItemRejectsInvalidIdentityWire(t *testing.T) {
	allowed := map[string]struct{}{"id": {}, "revision": {}, "email": {}}
	tests := []struct {
		name string
		item string
	}{
		{"missing id", `{"revision":"1"}`},
		{"empty id", `{"id":"","revision":"1"}`},
		{"non uuid id", `{"id":"not-a-uuid","revision":"1"}`},
		{"missing revision", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003"}`},
		{"null revision", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":null}`},
		{"numeric revision", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":1}`},
		{"empty revision", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":""}`},
		{"zero revision", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"0"}`},
		{"negative revision", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"-1"}`},
		{"negative zero", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"-0"}`},
		{"leading zero", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"01"}`},
		{"exponent", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"1e0"}`},
		{"overflow", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"9223372036854775808"}`},
		{"unknown field", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"1","unknown":"x"}`},
		{"secret field", `{"id":"018f47f4-7b1b-7cc0-8000-000000000003","revision":"1","password_hash":"x"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if projected, valid := projectSystemResourceQueryItem(
				json.RawMessage(test.item), allowed,
			); valid || projected != nil {
				t.Fatalf("invalid item was projected: %s", projected)
			}
		})
	}
}
