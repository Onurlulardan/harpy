package published

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func cloneParameterFieldType(fieldType api.FieldType) (api.FieldType, error) {
	kind, err := fieldType.Kind()
	if err != nil {
		return api.FieldType{}, fmt.Errorf("published parameter type is invalid")
	}
	switch kind {
	case api.LogicalTypeText:
		profile, found, accessorErr := fieldType.Text()
		if accessorErr != nil || !found {
			break
		}
		return api.NewTextFieldType(profile)
	case api.LogicalTypeInteger:
		profile, found, accessorErr := fieldType.Integer()
		if accessorErr != nil || !found {
			break
		}
		return api.NewIntegerFieldType(profile)
	case api.LogicalTypeDecimal:
		profile, found, accessorErr := fieldType.Decimal()
		if accessorErr != nil || !found {
			break
		}
		return api.NewDecimalFieldType(profile)
	case api.LogicalTypeBoolean:
		return api.NewBooleanFieldType()
	case api.LogicalTypeDate:
		return api.NewDateFieldType()
	case api.LogicalTypeTimestamp:
		return api.NewTimestampFieldType()
	case api.LogicalTypeUUID:
		return api.NewUUIDFieldType()
	case api.LogicalTypeEnum:
		profile, found, accessorErr := fieldType.Enum()
		if accessorErr != nil || !found {
			break
		}
		return api.NewEnumFieldType(profile)
	case api.LogicalTypeRelation:
		return api.NewRelationFieldType()
	case api.LogicalTypeTextList:
		profile, found, accessorErr := fieldType.Text()
		if accessorErr != nil || !found {
			break
		}
		return api.NewTextListFieldType(profile)
	case api.LogicalTypeUUIDList:
		return api.NewUUIDListFieldType()
	case api.LogicalTypeJSON:
		return api.NewJSONFieldType()
	}
	return api.FieldType{}, fmt.Errorf("published parameter type is invalid")
}
