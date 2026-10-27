package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

func TestCatalogContractValuesAreRedactedAndRejectDirectJSON(t *testing.T) {
	set, rootCatalog, _, _, _ := mustCatalogSetFixture(t)
	field := rootCatalog.fields[0]
	filter := field.capabilities.filters[0]
	values := []any{
		filter, field.capabilities, field, rootCatalog.capabilities, rootCatalog, set,
	}
	for _, value := range values {
		for _, formatted := range []string{
			fmt.Sprint(value), fmt.Sprintf("%s", value),
			fmt.Sprintf("%q", value), fmt.Sprintf("%#v", value),
			fmt.Sprintf("%x", value),
		} {
			if formatted != queryValueRedaction {
				t.Fatalf("%T formatting leaked catalog data: %q", value, formatted)
			}
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		} else if strings.Contains(err.Error(), "root_records") ||
			strings.Contains(err.Error(), rootResourceID) {
			t.Fatalf("%T marshal error leaked catalog data: %v", value, err)
		}
	}
}

func TestFieldDefinitionZeroTamperAndDeepCloneSafety(t *testing.T) {
	fieldType := mustEnumFieldType(t, true)
	token, _ := sqlidentifier.MintDynamic("fld_status")
	storage, _ := NewFieldStorage(token, nil, nil)
	capabilities, _ := NewFieldCapabilities(fieldType, FieldCapabilitiesInput{
		Selectable: true,
		Filters:    []FilterCapabilityInput{{Operator: OperatorEq, CountSafe: true}},
		Aggregates: []AggregateFunction{AggregateFunctionMin},
	})
	field, err := NewFieldDefinition(
		"b0000000-0000-4000-8000-000000000001", "status",
		fieldType, storage, capabilities,
	)
	if err != nil {
		t.Fatal(err)
	}
	originalState := fieldType.state.(enumFieldState)
	originalState.profile.values[0].code = "mutated"
	if field.Validate() != nil {
		t.Fatal("field definition retained input FieldType storage")
	}
	returnedType, _ := field.Type()
	returnedState := returnedType.state.(enumFieldState)
	returnedState.profile.values[0].code = "mutated_again"
	again, err := field.Type()
	if err != nil || again.Validate() != nil {
		t.Fatal("field definition exposed FieldType storage")
	}
	missingID := field.clone()
	missingID.id = uuid.Nil
	badCapabilities := field.clone()
	badCapabilities.capabilities.filters[0] = FilterCapability{}
	for _, invalid := range []FieldDefinition{{}, missingID, badCapabilities} {
		if invalid.Validate() == nil {
			t.Fatal("invalid field definition validated")
		}
		checks := []func() error{
			func() error { _, err := invalid.ID(); return err },
			func() error { _, err := invalid.PublicName(); return err },
			func() error { _, err := invalid.Type(); return err },
			func() error { _, err := invalid.Storage(); return err },
			func() error { _, err := invalid.Capabilities(); return err },
		}
		for _, check := range checks {
			if check() == nil {
				t.Fatal("invalid field definition accessor succeeded")
			}
		}
	}
}

func TestCatalogConstructionErrorsDoNotExposeInputs(t *testing.T) {
	valid := mustUUIDCatalogField(
		t, rootStableField, "record_id", "fld_record_id", true, true, 36,
	)
	privateName := "private\x00field"
	_, err := NewFieldDefinition(
		rootStableField, privateName, valid.fieldType, valid.storage, valid.capabilities,
	)
	if err == nil || strings.Contains(err.Error(), privateName) ||
		strings.Contains(err.Error(), rootStableField) {
		t.Fatalf("field definition error leaked input: %v", err)
	}
}
