package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestFieldCapabilitiesZeroTamperAndFormattingFailClosed(t *testing.T) {
	valid, _ := NewFieldCapabilities(mustCasefoldTextFieldType(t), FieldCapabilitiesInput{
		Filters: []FilterCapabilityInput{{Operator: OperatorEq}},
	})
	unsorted := valid.clone()
	second, _ := newFilterCapability(FilterCapabilityInput{Operator: OperatorNe})
	unsorted.filters = []FilterCapability{second, valid.filters[0]}
	wrongKind := valid.clone()
	wrongKind.logicalType = LogicalTypeJSON
	badFilter := valid.clone()
	badFilter.filters[0] = FilterCapability{}
	for _, capabilities := range []FieldCapabilities{{}, unsorted, wrongKind, badFilter} {
		if capabilities.Validate() == nil {
			t.Fatal("invalid field capabilities validated")
		}
		if _, err := capabilities.Filters(); err == nil {
			t.Fatal("invalid field capabilities accessor succeeded")
		}
	}
	for _, value := range []any{valid, valid.filters[0]} {
		if fmt.Sprint(value) != queryValueRedaction ||
			fmt.Sprintf("%#v", value) != queryValueRedaction {
			t.Fatalf("%T formatting was not redacted", value)
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}
