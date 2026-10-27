package authorization

func (fact BindingFact) validUserShape() bool {
	return fact.PermissionMatchKind == "exact" && fact.AssignmentID == nil &&
		fact.AssignmentState == "" && fact.RoleID == nil && fact.RoleStatus == "" &&
		fact.RoleScopeKind == "" && fact.MembershipID == nil &&
		fact.MembershipState == "" && fact.OrganizationID == nil &&
		fact.OrganizationStatus == "" && fact.noPolicy()
}

func (fact BindingFact) validRoleShape() bool {
	if !validUUIDPointer(fact.AssignmentID) || !validUUIDPointer(fact.RoleID) ||
		!validAssignmentState(fact.AssignmentState) || !validRoleStatus(fact.RoleStatus) {
		return false
	}
	switch fact.RoleScopeKind {
	case "tenant":
		if fact.MembershipID != nil || fact.MembershipState != "" ||
			fact.OrganizationID != nil || fact.OrganizationStatus != "" {
			return false
		}
	case "organization":
		if !validUUIDPointer(fact.MembershipID) ||
			!validMembershipState(fact.MembershipState) ||
			!validUUIDPointer(fact.OrganizationID) ||
			!validOrganizationStatus(fact.OrganizationStatus) {
			return false
		}
	default:
		return false
	}
	return fact.validPolicyShape()
}

func (fact BindingFact) validOrganizationShape() bool {
	return fact.PermissionMatchKind == "exact" && fact.AssignmentID == nil &&
		fact.AssignmentState == "" && fact.RoleID == nil && fact.RoleStatus == "" &&
		fact.RoleScopeKind == "" && validUUIDPointer(fact.MembershipID) &&
		validMembershipState(fact.MembershipState) &&
		validUUIDPointer(fact.OrganizationID) &&
		validOrganizationStatus(fact.OrganizationStatus) && fact.noPolicy()
}

func (fact BindingFact) validPolicyShape() bool {
	if fact.PolicyRoleID == nil && fact.PolicyPermissionID == nil &&
		fact.PolicyBindingID == nil && fact.PolicyState == "" {
		return true
	}
	return validUUIDPointer(fact.PolicyRoleID) &&
		validUUIDPointer(fact.PolicyPermissionID) &&
		validUUIDPointer(fact.PolicyBindingID) && fact.PolicyState == "active"
}

func (fact BindingFact) noPolicy() bool {
	return fact.PolicyRoleID == nil && fact.PolicyPermissionID == nil &&
		fact.PolicyBindingID == nil && fact.PolicyState == ""
}

func validAssignmentState(value string) bool {
	return value == "active" || value == "suspended" || value == "revoked"
}

func validRoleStatus(value string) bool { return value == "active" || value == "retired" }

func validMembershipState(value string) bool {
	return value == "active" || value == "suspended" || value == "tombstoned"
}

func validOrganizationStatus(value string) bool {
	return value == "active" || value == "suspended" || value == "tombstoned"
}
