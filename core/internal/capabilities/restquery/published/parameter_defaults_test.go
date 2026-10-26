package published

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestParameterManifestRequiresExactCanonicalDefaults(t *testing.T) {
	integerType := testIntegerFieldType(t, 0, 10)
	decimalType := testDecimalFieldType(t, 5, 2)
	tests := []struct {
		name       string
		fieldType  api.FieldType
		arity      ParameterArity
		required   bool
		defaults   []resolved.Value
		wantReject bool
	}{
		{"required default", integerType, ParameterAritySingle, true,
			[]resolved.Value{mustScalarValue(t, api.LogicalTypeInteger, "5")}, true},
		{"out of range", integerType, ParameterAritySingle, false,
			[]resolved.Value{mustScalarValue(t, api.LogicalTypeInteger, "11")}, true},
		{"wrong type", integerType, ParameterAritySingle, false,
			[]resolved.Value{mustScalarValue(t, api.LogicalTypeText, "5")}, true},
		{"decimal exact", decimalType, ParameterAritySingle, false,
			[]resolved.Value{mustScalarValue(t, api.LogicalTypeDecimal, "1.20")}, false},
		{"decimal needs normalization", decimalType, ParameterAritySingle, false,
			[]resolved.Value{mustScalarValue(t, api.LogicalTypeDecimal, "1.2")}, true},
		{"single has two", integerType, ParameterAritySingle, false,
			[]resolved.Value{
				mustScalarValue(t, api.LogicalTypeInteger, "1"),
				mustScalarValue(t, api.LogicalTypeInteger, "2"),
			}, true},
		{"list empty", integerType, ParameterArityList, false, []resolved.Value{}, true},
		{"list duplicate", integerType, ParameterArityList, false,
			[]resolved.Value{
				mustScalarValue(t, api.LogicalTypeInteger, "1"),
				mustScalarValue(t, api.LogicalTypeInteger, "1"),
			}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewParameterManifest([]ParameterInput{{
				Name: "value", Type: test.fieldType, Arity: test.arity,
				Required: test.required, Default: test.defaults,
			}}, 50)
			if (err != nil) != test.wantReject {
				t.Fatalf("error = %v, reject = %v", err, test.wantReject)
			}
		})
	}
}

func TestParameterManifestRestrictsPublishedFieldTypes(t *testing.T) {
	jsonType, err := api.NewJSONFieldType()
	if err != nil {
		t.Fatal(err)
	}
	profile, err := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	if err != nil {
		t.Fatal(err)
	}
	textListType, err := api.NewTextListFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []ParameterInput{
		{Name: "json_list", Type: jsonType, Arity: ParameterArityList},
		{Name: "collection", Type: textListType, Arity: ParameterAritySingle},
	} {
		if _, err := NewParameterManifest([]ParameterInput{input}, 50); err == nil {
			t.Fatalf("unsupported input accepted: %s", input.Name)
		}
	}
	if _, err := NewParameterManifest([]ParameterInput{{
		Name: "json_value", Type: jsonType, Arity: ParameterAritySingle,
	}}, 50); err != nil {
		t.Fatalf("single JSON parameter rejected: %v", err)
	}
}
