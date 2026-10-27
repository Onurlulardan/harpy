package published

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

func decodeParameterDefaults(
	raw json.RawMessage, fieldType api.FieldType, arity ParameterArity,
) ([]resolved.Value, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	values := []json.RawMessage{raw}
	if arity == ParameterArityList {
		if err := json.Unmarshal(raw, &values); err != nil || len(values) == 0 {
			return nil, fmt.Errorf("published parameter default is invalid")
		}
	}
	literals := make([]api.Literal, len(values))
	for index, value := range values {
		literal, err := api.NewJSONLiteral(bytes.TrimSpace(value))
		if err != nil {
			return nil, err
		}
		literals[index] = literal
	}
	operator := api.OperatorEq
	if arity == ParameterArityList {
		operator = api.OperatorIn
	}
	return resolver.ResolveLiterals(fieldType, operator, literals)
}
