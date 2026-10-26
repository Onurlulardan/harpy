package adoption

import "sort"

func sortedResources(snapshot Snapshot) []SourceResource {
	result := append([]SourceResource(nil), snapshot.Resources...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID.String() < result[j].ID.String() })
	return result
}

func sortedActions(snapshot Snapshot) []SourceAction {
	result := append([]SourceAction(nil), snapshot.Actions...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID.String() < result[j].ID.String() })
	return result
}

func sortedPermissions(snapshot Snapshot) []SourcePermission {
	result := append([]SourcePermission(nil), snapshot.Permissions...)
	for index := range result {
		result[index].ActionIDs = append(result[index].ActionIDs[:0:0], result[index].ActionIDs...)
		sortUUIDs(result[index].ActionIDs)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID.String() < result[j].ID.String() })
	return result
}

func sortedRequests(snapshot Snapshot) []SourcePermissionRequest {
	result := append([]SourcePermissionRequest(nil), snapshot.PermissionRequests...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID.String() < result[j].ID.String() })
	return result
}

func sortedInvitations(snapshot Snapshot) []SourceInvitation {
	result := append([]SourceInvitation(nil), snapshot.Invitations...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID.String() < result[j].ID.String() })
	return result
}

func sortedProfileFields(snapshot Snapshot) []SourceProfileField {
	result := append([]SourceProfileField(nil), snapshot.ProfileFields...)
	sort.Slice(result, func(i, j int) bool {
		return profileKeyText(profileKey(result[i].ResourceKind, result[i].Key)) <
			profileKeyText(profileKey(result[j].ResourceKind, result[j].Key))
	})
	return result
}

func sortedLayoutFields(snapshot Snapshot) []SourceLayoutField {
	result := append([]SourceLayoutField(nil), snapshot.LayoutFields...)
	sort.Slice(result, func(i, j int) bool {
		left := result[i].ResourceKind + result[i].SectionKey + result[i].FieldKey
		right := result[j].ResourceKind + result[j].SectionKey + result[j].FieldKey
		return left < right
	})
	return result
}
