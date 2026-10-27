package published

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func ValidateTemplate(
	document []byte, manifest ParameterManifest, decoder api.POSTDecoder,
) error {
	definitions, err := manifest.Parameters()
	if err != nil {
		return err
	}
	parameters := make([]boundParameter, len(definitions))
	for index, definition := range definitions {
		values, present, defaultErr := definition.Default()
		if defaultErr != nil {
			return defaultErr
		}
		if !present {
			value, sampleErr := sampleParameterValue(definition)
			if sampleErr != nil {
				return sampleErr
			}
			values = []resolved.Value{value}
		}
		parameters[index] = boundParameter{definition: definition, values: values}
	}
	bound, err := newBoundParameters(parameters, nil)
	if err != nil {
		return err
	}
	_, err = ResolveTemplate(document, bound, decoder)
	return err
}

func sampleParameterValue(definition ParameterDefinition) (resolved.Value, error) {
	fieldType, err := definition.Type()
	if err != nil {
		return resolved.Value{}, err
	}
	kind, err := fieldType.Kind()
	if err != nil {
		return resolved.Value{}, err
	}
	switch kind {
	case api.LogicalTypeText:
		return resolved.NewScalarValue(kind, "sample")
	case api.LogicalTypeInteger:
		profile, _, _ := fieldType.Integer()
		minimum, minimumErr := profile.Minimum()
		if minimumErr != nil {
			return resolved.Value{}, minimumErr
		}
		return resolved.NewScalarValue(kind, strconv.FormatInt(minimum, 10))
	case api.LogicalTypeDecimal:
		profile, _, _ := fieldType.Decimal()
		scale, scaleErr := profile.Scale()
		if scaleErr != nil {
			return resolved.Value{}, scaleErr
		}
		value := "0"
		if scale > 0 {
			value += "." + strings.Repeat("0", scale)
		}
		return resolved.NewScalarValue(kind, value)
	case api.LogicalTypeBoolean:
		return resolved.NewBooleanValue(false)
	case api.LogicalTypeDate:
		return resolved.NewScalarValue(kind, "2000-01-01")
	case api.LogicalTypeTimestamp:
		return resolved.NewScalarValue(kind, "2000-01-01T00:00:00Z")
	case api.LogicalTypeUUID, api.LogicalTypeRelation:
		return resolved.NewScalarValue(kind, "018f0000-0000-7000-8000-000000000001")
	case api.LogicalTypeEnum:
		profile, _, _ := fieldType.Enum()
		values, valuesErr := profile.Values()
		if valuesErr != nil || len(values) == 0 {
			return resolved.Value{}, fmt.Errorf("published enum sample is invalid")
		}
		code, codeErr := values[0].Code()
		if codeErr != nil {
			return resolved.Value{}, codeErr
		}
		return resolved.NewScalarValue(kind, code)
	case api.LogicalTypeJSON:
		return resolved.NewJSONValue([]byte(`{}`))
	default:
		return resolved.Value{}, fmt.Errorf("published parameter sample is unsupported")
	}
}
