package cursor

import (
	"encoding/json"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func encodeWireSorts(sorts []SortDefinition) []wireSort {
	result := make([]wireSort, len(sorts))
	for index, sort := range sorts {
		result[index] = wireSort{
			FieldID: sort.fieldID.String(), Direction: sort.direction.String(),
			LogicalType: sort.logicalType.String(),
		}
	}
	return result
}

func encodeWireLastValues(values []LastValue) ([]wireLastValue, error) {
	result := make([]wireLastValue, len(values))
	for index, value := range values {
		if value.null {
			result[index] = wireLastValue{Null: true, Value: json.RawMessage("null")}
			continue
		}
		kind, err := value.value.Kind()
		if err != nil {
			return nil, errEncodeFailed
		}
		if kind == api.LogicalTypeBoolean {
			boolean, ok, booleanErr := value.value.Boolean()
			if booleanErr != nil || !ok {
				return nil, errEncodeFailed
			}
			result[index].Value = json.RawMessage(strconv.FormatBool(boolean))
			continue
		}
		scalar, ok, scalarErr := value.value.Scalar()
		if scalarErr != nil || !ok {
			return nil, errEncodeFailed
		}
		encoded, marshalErr := json.Marshal(scalar)
		if marshalErr != nil {
			return nil, errEncodeFailed
		}
		result[index].Value = encoded
	}
	return result, nil
}
