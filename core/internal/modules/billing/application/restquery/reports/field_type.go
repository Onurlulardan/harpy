package reports

import (
	"math"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func fieldType(spec fieldSpec) (api.FieldType, error) {
	switch spec.Kind {
	case "uuid":
		return api.NewUUIDFieldType()
	case "boolean":
		return api.NewBooleanFieldType()
	case "timestamp":
		return api.NewTimestampFieldType()
	case "text":
		profile, err := api.NewTextProfile(
			api.TextCasefoldDisabled, api.TextAccentDisabled,
		)
		if err != nil {
			return api.FieldType{}, ErrContract
		}
		return api.NewTextFieldType(profile)
	case "integer":
		profile, err := api.NewIntegerProfile(math.MinInt64, math.MaxInt64)
		if err != nil {
			return api.FieldType{}, ErrContract
		}
		return api.NewIntegerFieldType(profile)
	case "json":
		return api.NewJSONFieldType()
	default:
		return api.FieldType{}, ErrContract
	}
}
