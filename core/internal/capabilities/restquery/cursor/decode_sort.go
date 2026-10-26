package cursor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func decodeWireSorts(values []wireSort) ([]decodedSort, bool) {
	if len(values) < 1 || len(values) > maximumSorts {
		return nil, false
	}
	result := make([]decodedSort, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for index, value := range values {
		fieldID, fieldOK := parseCanonicalUUID(value.FieldID)
		direction, directionOK := api.ParseDirection(value.Direction)
		logicalType, typeOK := api.ParseLogicalType(value.LogicalType)
		if !fieldOK || !directionOK || !typeOK || !sortableLogicalType(logicalType) {
			return nil, false
		}
		if _, duplicate := seen[fieldID]; duplicate {
			return nil, false
		}
		seen[fieldID] = struct{}{}
		result[index] = decodedSort{
			fieldID: fieldID, direction: direction, logicalType: logicalType,
		}
	}
	return result, true
}
