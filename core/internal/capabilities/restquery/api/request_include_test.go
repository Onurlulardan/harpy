package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRequestIncludeBuildsImmutableRecursiveValue(t *testing.T) {
	where := mustPredicate(t)
	limit := 20
	child := mustRequestInclude(t, RequestIncludeInput{
		Relation: "items", Select: []string{"id"},
	})
	input := RequestIncludeInput{
		Relation: "customer", Select: []string{"id", "name"}, Where: &where,
		Sort: []RequestSort{mustRequestSort(t, "name")}, Limit: &limit,
		Include: []RequestInclude{child},
	}
	include := mustRequestInclude(t, input)
	input.Select[0] = "changed"
	input.Sort[0] = RequestSort{}
	input.Include[0] = RequestInclude{}
	limit = 99
	if include.Relation() != "customer" || include.depth != 2 || include.edges != 2 {
		t.Fatalf("unexpected include: %#v", include)
	}
	if actual, found := include.Limit(); !found || actual != 20 {
		t.Fatalf("limit = %d, found = %v", actual, found)
	}
	selectFields := include.Select()
	selectFields[0] = "changed"
	children := include.Includes()
	children[0] = RequestInclude{}
	if include.Select()[0] != "id" || include.Includes()[0].Relation() != "items" {
		t.Fatal("include exposed mutable storage")
	}
	if _, found := include.Where(); !found {
		t.Fatal("include lost optional filter")
	}
}

func TestRequestIncludeValidatesNamesSortAndLimit(t *testing.T) {
	zero, over := 0, maximumIncludeLimit+1
	for _, input := range []RequestIncludeInput{
		{Relation: ""},
		{Relation: "e\u0301"},
		{Relation: "relation\x00name"},
		{Relation: "relation", Select: []string{"field", "field"}},
		{Relation: "relation", Sort: []RequestSort{{}}},
		{Relation: "relation", Limit: &zero},
		{Relation: "relation", Limit: &over},
	} {
		if _, err := NewRequestInclude(input); err == nil {
			t.Fatalf("invalid include accepted: %#v", input)
		}
	}
}

func TestRequestIncludeIsRedactedAndCannotMarshal(t *testing.T) {
	input := RequestIncludeInput{Relation: "secret_relation", Select: []string{"secret_field"}}
	include := mustRequestInclude(t, input)
	for _, value := range []string{
		fmt.Sprint(include), fmt.Sprintf("%#v", include),
		fmt.Sprint(input), fmt.Sprintf("%#v", input),
	} {
		if value != queryValueRedaction {
			t.Fatalf("include formatting leaked: %q", value)
		}
	}
	for _, value := range []any{include, input} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}

func mustRequestInclude(t *testing.T, input RequestIncludeInput) RequestInclude {
	t.Helper()
	include, err := NewRequestInclude(input)
	if err != nil {
		t.Fatal(err)
	}
	return include
}
