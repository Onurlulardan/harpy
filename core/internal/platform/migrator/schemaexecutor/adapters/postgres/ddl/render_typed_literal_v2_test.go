package ddl

import (
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func TestRenderTypedLiteralV2QuotesWithoutInterpolation(t *testing.T) {
	literal := schemav2.TypedLiteral{StorageType: schemav2.StorageText,
		Value: []byte(`"x'); DROP TABLE records; --"`)}
	result, err := RenderTypedLiteralV2(literal)
	if err != nil {
		t.Fatal(err)
	}
	want := `'x''); DROP TABLE records; --'::text`
	if result != want {
		t.Fatalf("literal = %s, want %s", result, want)
	}
}

func TestRenderTypedLiteralV2TypesArraysAndJSON(t *testing.T) {
	tests := []struct {
		literal schemav2.TypedLiteral
		want    string
	}{
		{schemav2.TypedLiteral{StorageType: schemav2.StorageUUIDList,
			Value: []byte(`["018f47f4-7b1b-7cc0-8000-000000000001"]`)},
			`ARRAY['018f47f4-7b1b-7cc0-8000-000000000001']::uuid[]`},
		{schemav2.TypedLiteral{StorageType: schemav2.StorageJSON,
			Value: []byte(`{"name":"O'Reilly"}`)}, `'{"name":"O''Reilly"}'::jsonb`},
	}
	for _, test := range tests {
		result, err := RenderTypedLiteralV2(test.literal)
		if err != nil || result != test.want {
			t.Fatalf("literal = %s, err = %v, want %s", result, err, test.want)
		}
	}
}
