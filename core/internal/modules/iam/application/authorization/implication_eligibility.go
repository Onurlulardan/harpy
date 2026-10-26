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

func implicationReason(kind string, path SubjectPath) (Reason, int, bool) {
	switch kind {
	case "v1_form_parent_v1":
		return compatibilityReason(path)
	case "forms_record_operation_v1":
		return formsRecordOperationReason(path)
	default:
		return "", 0, false
	}
}

func compatibilityReason(path SubjectPath) (Reason, int, bool) {
	switch path {
	case SubjectUser:
		return ReasonUserCompatibility, 60, true
	case SubjectRole:
		return ReasonRoleCompatibility, 70, true
	case SubjectOrganization:
		return ReasonOrganizationCompatibility, 80, true
	default:
		return "", 0, false
	}
}

func formsRecordOperationReason(path SubjectPath) (Reason, int, bool) {
	switch path {
	case SubjectUser:
		return ReasonUserFormsRecordOperation, 60, true
	case SubjectRole:
		return ReasonRoleFormsRecordOperation, 70, true
	case SubjectOrganization:
		return ReasonOrganizationFormsRecordOperation, 80, true
	default:
		return "", 0, false
	}
}
