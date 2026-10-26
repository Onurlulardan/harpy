package published

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestParameterManifestOwnsDefinitionsAndDefaults(t *testing.T) {
	defaults := []resolved.Value{
		mustScalarValue(t, api.LogicalTypeInteger, "5"),
	}
	manifest, err := NewParameterManifest([]ParameterInput{{
		Name: "minimum", Type: testIntegerFieldType(t, 0, 10),
		Arity: ParameterAritySingle, Default: defaults,
	}}, 100)
	if err != nil {
		t.Fatal(err)
	}
	defaults[0] = mustScalarValue(t, api.LogicalTypeInteger, "7")
	definition, found, lookupErr := manifest.Lookup("minimum")
	if lookupErr != nil || !found {
		t.Fatal("manifest lookup failed")
	}
	stored, present, defaultErr := definition.Default()
	if defaultErr != nil || !present || scalarValue(stored[0]) != "5" {
		t.Fatal("manifest retained caller default storage")
	}
	stored[0] = mustScalarValue(t, api.LogicalTypeInteger, "9")
	again, _, _ := definition.Default()
	if scalarValue(again[0]) != "5" {
		t.Fatal("definition exposed default storage")
	}
	parameters, parametersErr := manifest.Parameters()
	if parametersErr != nil || len(parameters) != 1 {
		t.Fatal("manifest parameters unavailable")
	}
	parameters[0] = ParameterDefinition{}
	if _, found, err := manifest.Lookup("minimum"); err != nil || !found {
		t.Fatal("manifest exposed parameter storage")
	}
	maximum, maximumErr := manifest.MaximumPageSize()
	if maximumErr != nil || maximum != 100 {
		t.Fatal("manifest page maximum changed")
	}
}

func TestParameterManifestRejectsInvalidNamesDuplicatesAndMaximum(t *testing.T) {
	fieldType := testTextFieldType(t)
	valid63 := "a" + strings.Repeat("0", 62)
	for _, name := range []string{"a", valid63} {
		if _, err := NewParameterManifest([]ParameterInput{{
			Name: name, Type: fieldType, Arity: ParameterAritySingle,
		}}, 1); err != nil {
			t.Fatalf("valid name rejected: %q", name)
		}
	}
	for _, name := range []string{
		"", "A", "0name", "bad-name", "é", "_included",
		"a" + strings.Repeat("0", 63),
	} {
		if _, err := NewParameterManifest([]ParameterInput{{
			Name: name, Type: fieldType, Arity: ParameterAritySingle,
		}}, 50); err == nil {
			t.Fatalf("invalid name accepted: %q", name)
		}
	}
	duplicate := []ParameterInput{
		{Name: "same", Type: fieldType, Arity: ParameterAritySingle},
		{Name: "same", Type: fieldType, Arity: ParameterAritySingle},
	}
	if _, err := NewParameterManifest(duplicate, 50); err == nil {
		t.Fatal("duplicate parameter accepted")
	}
	for _, maximum := range []int{0, 501} {
		if _, err := NewParameterManifest(nil, maximum); err == nil {
			t.Fatal("invalid maximum page size accepted")
		}
	}
}

func TestParameterManifestEnforcesParameterLimit(t *testing.T) {
	fieldType := testTextFieldType(t)
	inputs := make([]ParameterInput, maximumPublishedParameters+1)
	for index := range inputs {
		inputs[index] = ParameterInput{
			Name: fmt.Sprintf("p%d", index), Type: fieldType,
			Arity: ParameterAritySingle,
		}
	}
	manifest, err := NewParameterManifest(inputs[:maximumPublishedParameters], 50)
	if err != nil {
		t.Fatalf("100 parameters rejected: %v", err)
	}
	if _, err := NewParameterManifest(inputs, 50); err == nil {
		t.Fatal("101 parameters accepted")
	}
	extra, err := NewParameterManifest(inputs[maximumPublishedParameters:], 50)
	if err != nil {
		t.Fatal(err)
	}
	manifest.parameters = append(manifest.parameters, extra.parameters[0])
	if err := manifest.Validate(); err == nil {
		t.Fatal("tampered 101-parameter manifest validated")
	}
}

func TestZeroParameterManifestIsInvalid(t *testing.T) {
	var manifest ParameterManifest
	if err := manifest.Validate(); err == nil {
		t.Fatal("zero manifest validated")
	}
	if _, err := manifest.Parameters(); err == nil {
		t.Fatal("zero manifest exposed parameters")
	}
	if _, _, err := manifest.Lookup("value"); err == nil {
		t.Fatal("zero manifest allowed lookup")
	}
	if _, err := manifest.MaximumPageSize(); err == nil {
		t.Fatal("zero manifest exposed page maximum")
	}
}

func TestEmptyParameterManifestIsValid(t *testing.T) {
	manifest, err := NewParameterManifest(nil, 50)
	if err != nil {
		t.Fatal(err)
	}
	parameters, err := manifest.Parameters()
	if err != nil || len(parameters) != 0 {
		t.Fatalf("parameters = %d, err = %v", len(parameters), err)
	}
}

func scalarValue(value resolved.Value) string {
	scalar, _, _ := value.Scalar()
	return scalar
}
