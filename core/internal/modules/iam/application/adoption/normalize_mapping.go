package adoption

import "sort"

func normalizedMappingForHash(source Mapping) Mapping {
	result := source
	result.Identity.Users = append([]UserMapping(nil), source.Identity.Users...)
	for index := range result.Identity.Users {
		assignments := append([]AssignmentMapping(nil), result.Identity.Users[index].Assignments...)
		sort.Slice(assignments, func(i, j int) bool {
			return assignments[i].RoleID.String() < assignments[j].RoleID.String()
		})
		result.Identity.Users[index].Assignments = assignments
	}
	sort.Slice(result.Identity.Users, func(i, j int) bool {
		return result.Identity.Users[i].UserID.String() < result.Identity.Users[j].UserID.String()
	})
	result.Resources = append([]ResourceMapping(nil), source.Resources...)
	sort.Slice(result.Resources, func(i, j int) bool {
		return result.Resources[i].SourceID.String() < result.Resources[j].SourceID.String()
	})
	result.Actions = append([]ActionMapping(nil), source.Actions...)
	sort.Slice(result.Actions, func(i, j int) bool {
		return result.Actions[i].SourceID.String() < result.Actions[j].SourceID.String()
	})
	result.Permissions = append([]PermissionMapping(nil), source.Permissions...)
	sort.Slice(result.Permissions, func(i, j int) bool {
		left := permissionMappingKey(result.Permissions[i].SourcePermissionID,
			result.Permissions[i].ResourceID, result.Permissions[i].ActionID)
		right := permissionMappingKey(result.Permissions[j].SourcePermissionID,
			result.Permissions[j].ResourceID, result.Permissions[j].ActionID)
		return left < right
	})
	normalizeExtensionMappings(&result, source)
	return result
}
