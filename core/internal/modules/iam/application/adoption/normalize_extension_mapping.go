package adoption

import "sort"

func normalizeExtensionMappings(result *Mapping, source Mapping) {
	result.PermissionRequests = append([]PermissionRequestMapping(nil), source.PermissionRequests...)
	sort.Slice(result.PermissionRequests, func(i, j int) bool {
		return result.PermissionRequests[i].SourceRequestID.String() <
			result.PermissionRequests[j].SourceRequestID.String()
	})
	result.Invitations = append([]InvitationMapping(nil), source.Invitations...)
	sort.Slice(result.Invitations, func(i, j int) bool {
		return result.Invitations[i].SourceInvitationID.String() <
			result.Invitations[j].SourceInvitationID.String()
	})
	result.ProfileSchemas = append([]ProfileSchemaMapping(nil), source.ProfileSchemas...)
	sort.Slice(result.ProfileSchemas, func(i, j int) bool {
		return result.ProfileSchemas[i].ResourceKind < result.ProfileSchemas[j].ResourceKind
	})
	result.ProfileFields = append([]ProfileFieldMapping(nil), source.ProfileFields...)
	sort.Slice(result.ProfileFields, func(i, j int) bool {
		left := profileKeyText(profileKey(result.ProfileFields[i].ResourceKind,
			result.ProfileFields[i].SourceKey))
		right := profileKeyText(profileKey(result.ProfileFields[j].ResourceKind,
			result.ProfileFields[j].SourceKey))
		return left < right
	})
	result.LayoutFields = append([]LayoutFieldMapping(nil), source.LayoutFields...)
	sort.Slice(result.LayoutFields, func(i, j int) bool {
		left := profileKeyText(profileKey(result.LayoutFields[i].ResourceKind,
			result.LayoutFields[i].SourceKey))
		right := profileKeyText(profileKey(result.LayoutFields[j].ResourceKind,
			result.LayoutFields[j].SourceKey))
		return left < right
	})
}
