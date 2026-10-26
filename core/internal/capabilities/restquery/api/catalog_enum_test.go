package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestLogicalTypeCanonicalClosedSet(t *testing.T) {
	tests := map[string]LogicalType{
		"text": LogicalTypeText, "integer": LogicalTypeInteger,
		"decimal": LogicalTypeDecimal, "boolean": LogicalTypeBoolean,
		"date": LogicalTypeDate, "timestamp": LogicalTypeTimestamp,
		"uuid": LogicalTypeUUID, "enum": LogicalTypeEnum,
		"relation": LogicalTypeRelation, "text_list": LogicalTypeTextList,
		"uuid_list": LogicalTypeUUIDList, "json": LogicalTypeJSON,
	}
	for name, want := range tests {
		actual, found := ParseLogicalType(name)
		if !found || actual != want || actual.String() != name ||
			!actual.Valid() || actual.Validate() != nil {
			t.Fatalf("logical type %q did not round trip", name)
		}
		if fmt.Sprint(actual) != name || fmt.Sprintf("%#v", actual) != name {
			t.Fatalf("logical type %q formatting was not canonical", name)
		}
	}
	for _, name := range []string{"", "Text", "number", "uuid-list"} {
		if value, found := ParseLogicalType(name); found || value != 0 {
			t.Fatalf("unknown logical type accepted: %q", name)
		}
	}
	if LogicalType(0).Valid() || LogicalType(255).Validate() == nil {
		t.Fatal("invalid logical type validated")
	}
}

func TestRelationCardinalityCanonicalClosedSet(t *testing.T) {
	tests := map[string]RelationCardinality{
		"to_one":  RelationCardinalityToOne,
		"to_many": RelationCardinalityToMany,
	}
	for name, want := range tests {
		actual, found := ParseRelationCardinality(name)
		if !found || actual != want || actual.String() != name ||
			!actual.Valid() || actual.Validate() != nil {
			t.Fatalf("relation cardinality %q did not round trip", name)
		}
		if fmt.Sprint(actual) != name || fmt.Sprintf("%#v", actual) != name {
			t.Fatalf("cardinality %q formatting was not canonical", name)
		}
	}
	for _, name := range []string{"", "to-one", "one", "TO_MANY"} {
		if value, found := ParseRelationCardinality(name); found || value != 0 {
			t.Fatalf("unknown cardinality accepted: %q", name)
		}
	}
	if RelationCardinality(0).Valid() || RelationCardinality(255).Validate() == nil {
		t.Fatal("invalid relation cardinality validated")
	}
}

func TestCatalogEnumsCannotMarshalDirectly(t *testing.T) {
	for _, value := range []any{LogicalTypeText, RelationCardinalityToMany} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}
