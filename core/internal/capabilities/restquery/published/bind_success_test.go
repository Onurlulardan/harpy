package published

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestBindGETResolvesDefaultsListsAndPageImmutably(t *testing.T) {
	manifest := testBindingManifest(t)
	size, after := 10, "opaque-cursor"
	page, err := NewPageOverride(&size, &after)
	if err != nil {
		t.Fatal(err)
	}
	provided := map[string][]api.Literal{
		"count": {mustTextLiteral(t, "7")},
		"flags": {
			mustTextLiteral(t, "true"), mustTextLiteral(t, "false"),
			mustTextLiteral(t, "true"),
		},
	}
	bound, err := BindGET(manifest, provided, &page)
	if err != nil {
		t.Fatal(err)
	}
	if err := bound.Validate(); err != nil {
		t.Fatal(err)
	}
	names, err := bound.Names()
	if err != nil || !reflect.DeepEqual(names, []string{"count", "label", "flags"}) {
		t.Fatalf("names = %#v, err = %v", names, err)
	}
	requireBoundScalar(t, bound, "count", "7")
	requireBoundScalar(t, bound, "label", "fallback")
	flags, found, err := bound.Lookup("flags")
	if err != nil || !found || len(flags) != 2 {
		t.Fatalf("flags = %#v, found = %v, err = %v", flags, found, err)
	}
	first, firstFound, _ := flags[0].Boolean()
	second, secondFound, _ := flags[1].Boolean()
	if !firstFound || !secondFound || !first || second {
		t.Fatalf("flags changed: %v %v", first, second)
	}
	flags[0] = mustBooleanValue(t, false)
	again, _, _ := bound.Lookup("flags")
	owned, _, _ := again[0].Boolean()
	if !owned {
		t.Fatal("bound values exposed mutable storage")
	}
	boundPage, present, err := bound.Page()
	if err != nil || !present {
		t.Fatalf("page present = %v, err = %v", present, err)
	}
	pageSize, hasSize, _ := boundPage.Size()
	pageAfter, hasAfter, _ := boundPage.After()
	if !hasSize || pageSize != 10 || !hasAfter || pageAfter != after {
		t.Fatal("bound page changed")
	}
}

func TestBindEmptyManifestProducesValidEmptyBoundSet(t *testing.T) {
	manifest, err := NewParameterManifest(nil, 50)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := BindGET(manifest, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	names, err := bound.Names()
	if err != nil || len(names) != 0 {
		t.Fatalf("names = %#v, err = %v", names, err)
	}
	if _, found, err := bound.Lookup("missing"); err != nil || found {
		t.Fatalf("missing found = %v, err = %v", found, err)
	}
}

func testBindingManifest(t *testing.T) ParameterManifest {
	t.Helper()
	manifest, err := NewParameterManifest([]ParameterInput{
		{Name: "count", Type: testIntegerFieldType(t, 0, 10),
			Arity: ParameterAritySingle, Required: true},
		{Name: "label", Type: testTextFieldType(t), Arity: ParameterAritySingle,
			Default: []resolved.Value{mustScalarValue(t, api.LogicalTypeText, "fallback")}},
		{Name: "flags", Type: testBooleanFieldType(t), Arity: ParameterArityList},
		{Name: "omitted", Type: testTextFieldType(t), Arity: ParameterAritySingle},
	}, 20)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func requireBoundScalar(
	t *testing.T, bound BoundParameters, name, expected string,
) {
	t.Helper()
	values, found, err := bound.Lookup(name)
	if err != nil || !found || len(values) != 1 || scalarValue(values[0]) != expected {
		t.Fatalf("%s = %#v, found = %v, err = %v", name, values, found, err)
	}
}
