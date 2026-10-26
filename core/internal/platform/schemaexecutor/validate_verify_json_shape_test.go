package schemaexecutor

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestArgumentPropertiesRejectNilObject(t *testing.T) {
	if err := validateArgumentProperties(schemav1.CreateFormTable, nil); err == nil {
		t.Fatal("expected null arguments rejection")
	}
	if err := validateArgumentProperties(
		schemav1.CreateFormTable, schemav1.OperationArguments{},
	); err != nil {
		t.Fatalf("validate empty object: %v", err)
	}
}

func TestVerifyNestedPropertiesRejectMissingBoolean(t *testing.T) {
	arguments := decodeOperationArguments(t, `{
		"checks":[],
		"fields":[{"field_id":"x","generation":1,"storage_type":"text"}],
		"relations":[],"indexes":[],"constraints":[]
	}`)
	if err := validateArgumentProperties(schemav1.VerifyStorage, arguments); err == nil {
		t.Fatal("expected missing nullable rejection")
	}
}

func TestVerifyNestedPropertiesRejectNullAndUnknown(t *testing.T) {
	tests := []string{
		`{"checks":[],"fields":[],"relations":[],"indexes":[` +
			`{"index_id":"x","fields":[],"method":"btree","unique":null,` +
			`"active_records_only":false}],"constraints":[]}`,
		`{"checks":[],"fields":[],"relations":[],"indexes":[],"constraints":[` +
			`{"constraint_id":"x","kind":"unique","fields":[],"extra":true}]}`,
		`{"checks":[],"fields":[],"relations":[],"indexes":[` +
			`{"index_id":"x","fields":[{"field_id":"x"}],"method":"btree",` +
			`"unique":false,"active_records_only":false}],"constraints":[]}`,
	}
	for _, document := range tests {
		arguments := decodeOperationArguments(t, document)
		if err := validateArgumentProperties(schemav1.VerifyStorage, arguments); err == nil {
			t.Fatalf("expected nested shape rejection for %s", document)
		}
	}
}

func TestVerifyPropertyContractsAlignWithJSONTypes(t *testing.T) {
	tests := []struct {
		value    any
		contract argumentPropertyContract
	}{
		{schemav1.VerifyFieldContract{}, verifyFieldPropertyContract},
		{schemav1.RelationArguments{}, verifyRelationPropertyContract},
		{schemav1.VerifyIndexContract{}, verifyIndexPropertyContract},
		{schemav1.VerifyConstraintContract{}, verifyConstraintPropertyContract},
		{schemav1.FieldReference{}, fieldReferencePropertyContract},
	}
	for _, test := range tests {
		required, optional := jsonProperties(reflect.TypeOf(test.value))
		wantRequired := slices.Clone(test.contract.required)
		wantOptional := slices.Clone(test.contract.optional)
		slices.Sort(wantRequired)
		slices.Sort(wantOptional)
		if !slices.Equal(required, wantRequired) || !slices.Equal(optional, wantOptional) {
			t.Fatalf("property contract drift for %T", test.value)
		}
	}
}

func decodeOperationArguments(t *testing.T, document string) schemav1.OperationArguments {
	t.Helper()
	var result schemav1.OperationArguments
	if err := json.Unmarshal([]byte(document), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func jsonProperties(value reflect.Type) ([]string, []string) {
	var required, optional []string
	for index := 0; index < value.NumField(); index++ {
		tag := value.Field(index).Tag.Get("json")
		parts := strings.Split(tag, ",")
		if slices.Contains(parts[1:], "omitempty") {
			optional = append(optional, parts[0])
		} else {
			required = append(required, parts[0])
		}
	}
	slices.Sort(required)
	slices.Sort(optional)
	return required, optional
}
