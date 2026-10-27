package authorization

func eligibleImplication(
	fact BindingFact,
	edge ImplicationFact,
) (Reason, int, bool) {
	if fact.PermissionID != edge.SourcePermissionID ||
		fact.PermissionMatchKind != "exact" || !bindingPathActive(fact) ||
		fact.PermissionState != "active" || fact.ResourceState != "active" ||
		fact.ActionState != "active" ||
		!sameUUID(fact.PermissionResource, edge.SourceResourceID) ||
		!sameUUID(fact.PermissionAction, edge.SourceActionID) {
		return "", 0, false
	}
	return implicationReason(edge.Kind, fact.Path)
}

func implicationReason(string, SubjectPath) (Reason, int, bool) {
	return "", 0, false
}
