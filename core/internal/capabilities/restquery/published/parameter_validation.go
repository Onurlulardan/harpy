package published

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (manifest ParameterManifest) Validate() error {
	if !manifest.validated || manifest.parameters == nil ||
		len(manifest.parameters) > maximumPublishedParameters ||
		manifest.maximumPageSize < 1 || manifest.maximumPageSize > 500 {
		return fmt.Errorf("published parameter manifest is invalid")
	}
	seen := make(map[string]struct{}, len(manifest.parameters))
	for _, definition := range manifest.parameters {
		if definition.Validate() != nil {
			return fmt.Errorf("published parameter manifest is invalid")
		}
		if _, duplicate := seen[definition.name]; duplicate {
			return fmt.Errorf("published parameter manifest is invalid")
		}
		seen[definition.name] = struct{}{}
	}
	return nil
}

func (definition ParameterDefinition) Validate() error {
	if !definition.validated || !validParameterName(definition.name) ||
		definition.fieldType.Validate() != nil || definition.arity.Validate() != nil ||
		definition.required && definition.defaults != nil ||
		!parameterTypeSupportsArity(definition.fieldType, definition.arity) {
		return fmt.Errorf("published parameter definition is invalid")
	}
	if definition.defaults != nil &&
		!validateResolvedParameterValues(
			definition.fieldType, definition.arity, definition.defaults,
		) {
		return fmt.Errorf("published parameter definition is invalid")
	}
	return nil
}

func validParameterName(name string) bool {
	if name == "_included" || len(name) < 1 || len(name) > 63 ||
		name[0] < 'a' || name[0] > 'z' {
		return false
	}
	for index := 1; index < len(name); index++ {
		character := name[index]
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func parameterTypeSupportsArity(fieldType api.FieldType, arity ParameterArity) bool {
	kind, err := fieldType.Kind()
	if err != nil {
		return false
	}
	switch kind {
	case api.LogicalTypeText, api.LogicalTypeInteger, api.LogicalTypeDecimal,
		api.LogicalTypeBoolean, api.LogicalTypeDate, api.LogicalTypeTimestamp,
		api.LogicalTypeUUID, api.LogicalTypeEnum, api.LogicalTypeRelation:
		return arity.Valid()
	case api.LogicalTypeJSON:
		return arity == ParameterAritySingle
	default:
		return false
	}
}
