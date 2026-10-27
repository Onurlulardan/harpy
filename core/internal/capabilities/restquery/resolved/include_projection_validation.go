package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func validIncludeProjections(
	projections []Projection,
	stable api.FieldDefinition,
) bool {
	if !validProjectionSet(projections, true) || !containsField(projections, stable) {
		return false
	}
	stableID, valid := fieldIdentifier(stable)
	if !valid {
		return false
	}
	hidden := 0
	for _, projection := range projections {
		if !projection.hidden {
			continue
		}
		fieldID, valid := fieldIdentifier(projection.field)
		if !valid || fieldID != stableID {
			return false
		}
		hidden++
	}
	return hidden <= 1
}
