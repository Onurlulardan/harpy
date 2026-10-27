package recordkernel

import "github.com/google/uuid"

func physicalCoverageValid(
	catalog Catalog, fields []Field, handles []physicalHandleDocument,
) bool {
	formCount := 0
	valueCounts := make(map[string]int)
	relationCounts := make(map[string]int)
	seenNames := make(map[string]struct{}, len(handles))
	for _, handle := range handles {
		if _, duplicate := seenNames[handle.PhysicalName]; duplicate {
			return false
		}
		seenNames[handle.PhysicalName] = struct{}{}
		switch handle.Kind {
		case physicalFormTable:
			formCount++
		case physicalFieldColumn:
			if handle.Variant == "value" {
				valueCounts[handle.LogicalID]++
			}
		case physicalRelationTable:
			if handle.FieldID == nil {
				return false
			}
			relationCounts[*handle.FieldID]++
		default:
			return false
		}
	}
	if formCount != 1 {
		return false
	}
	for _, field := range fields {
		if field.Role == RoleRelation {
			if field.Relation == nil || relationCounts[field.ID.String()] != 1 {
				return false
			}
			continue
		}
		if field.Relation != nil || valueCounts[field.ID.String()] != 1 {
			return false
		}
	}
	return catalog.FormID() != uuid.Nil
}
