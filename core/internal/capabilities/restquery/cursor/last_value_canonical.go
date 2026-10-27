package cursor

import (
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func validateLastValues(sorts []SortDefinition, values []LastValue) error {
	if len(values) != len(sorts) {
		return errInvalidLastValue
	}
	for index := range values {
		if values[index].Validate() != nil {
			return errInvalidLastValue
		}
		if values[index].null {
			continue
		}
		kind, err := values[index].value.Kind()
		if err != nil || kind != sorts[index].logicalType {
			return errInvalidLastValue
		}
		canonical, err := canonicalValueBytes(values[index])
		if err != nil || len(canonical) > sorts[index].canonicalBound {
			return errInvalidLastValue
		}
	}
	return nil
}

func canonicalValueBytes(value LastValue) ([]byte, error) {
	kind, err := value.value.Kind()
	if err != nil {
		return nil, errInvalidLastValue
	}
	if kind == api.LogicalTypeBoolean {
		boolean, ok, booleanErr := value.value.Boolean()
		if booleanErr != nil || !ok {
			return nil, errInvalidLastValue
		}
		return []byte(strconv.FormatBool(boolean)), nil
	}
	scalar, ok, scalarErr := value.value.Scalar()
	if scalarErr != nil || !ok {
		return nil, errInvalidLastValue
	}
	return []byte(scalar), nil
}
