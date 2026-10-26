package schemaexecutor

import (
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func TestValidateTypedLiteralV2AcceptsCanonicalForms(t *testing.T) {
	values := []schemav2.TypedLiteral{
		{StorageType: schemav2.StorageText, Value: []byte(`"O'Reilly"`)},
		{StorageType: schemav2.StorageInteger, Value: []byte(`"-42"`)},
		{StorageType: schemav2.StorageDecimal, Value: []byte(`"12.30"`)},
		{StorageType: schemav2.StorageBoolean, Value: []byte(`true`)},
		{StorageType: schemav2.StorageDate, Value: []byte(`"2026-07-18"`)},
		{StorageType: schemav2.StorageTimestamp,
			Value: []byte(`"2026-07-18T12:13:14.123456Z"`)},
		{StorageType: schemav2.StorageUUID,
			Value: []byte(`"018f47f4-7b1b-7cc0-8000-000000000001"`)},
		{StorageType: schemav2.StorageTextList, Value: []byte(`["a","b"]`)},
		{StorageType: schemav2.StorageJSON, Value: []byte(`{"a":1,"b":true}`)},
	}
	for _, value := range values {
		value := value
		t.Run(string(value.StorageType), func(t *testing.T) {
			if err := ValidateTypedLiteralV2(&value); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestValidateTypedLiteralV2RejectsAmbiguousValues(t *testing.T) {
	values := []schemav2.TypedLiteral{
		{StorageType: schemav2.StorageText, Value: []byte("null")},
		{StorageType: schemav2.StorageInteger, Value: []byte(`"01"`)},
		{StorageType: schemav2.StorageDecimal, Value: []byte(`"-0.00"`)},
		{StorageType: schemav2.StorageTimestamp,
			Value: []byte(`"2026-07-18T12:13:14Z"`)},
		{StorageType: schemav2.StorageUUID,
			Value: []byte(`"018F47F4-7B1B-7CC0-8000-000000000001"`)},
		{StorageType: schemav2.StorageTextList, Value: []byte(`["a", "b"]`)},
		{StorageType: schemav2.StorageJSON, Value: []byte(`{"b":true,"a":1}`)},
		{StorageType: schemav2.StorageRelation,
			Value: []byte(`"018f47f4-7b1b-7cc0-8000-000000000001"`)},
	}
	for _, value := range values {
		if err := ValidateTypedLiteralV2(&value); err == nil {
			t.Fatalf("accepted invalid literal: %+v", value)
		}
	}
}
