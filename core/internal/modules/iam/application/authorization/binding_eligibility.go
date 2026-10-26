package authorization

import "github.com/google/uuid"

func eligibleBinding(
	fact BindingFact,
	resourceID uuid.UUID,
	actionID uuid.UUID,
) (Reason, int, bool) {
	if !bindingPathActive(fact) || fact.PermissionState != "active" {
		return "", 0, false
	}
	switch fact.PermissionMatchKind {
	case "exact":
		if fact.ResourceState != "active" || fact.ActionState != "active" ||
			!sameUUID(fact.PermissionResource, resourceID) ||
			!sameUUID(fact.PermissionAction, actionID) {
			return "", 0, false
		}
		return exactReason(fact.Path)
	case "resource_all_actions":
		if fact.Path != SubjectRole || fact.ResourceState != "active" ||
			!sameUUID(fact.PermissionResource, resourceID) || !protectedPolicyRole(fact) {
			return "", 0, false
		}
		return ReasonProtectedResourceActions, 40, true
	case "tenant_all":
		if fact.Path != SubjectRole || !protectedTenantAll(fact) {
			return "", 0, false
		}
		return ReasonProtectedTenantAll, 50, true
	default:
		return "", 0, false
	}
}

func bindingPathActive(fact BindingFact) bool {
	if !fact.valid() || fact.BindingState != "active" {
		return false
	}
	switch fact.Path {
	case SubjectUser:
		return true
	case SubjectRole:
		if fact.AssignmentState != "active" || fact.RoleStatus != "active" {
			return false
		}
		return fact.RoleScopeKind == "tenant" ||
			(fact.RoleScopeKind == "organization" &&
				fact.MembershipState == "active" && fact.OrganizationStatus == "active")
	case SubjectOrganization:
		return fact.MembershipState == "active" && fact.OrganizationStatus == "active"
	default:
		return false
	}
}

func exactReason(path SubjectPath) (Reason, int, bool) {
	switch path {
	case SubjectUser:
		return ReasonDirectUserExact, 10, true
	case SubjectRole:
		return ReasonRoleExact, 20, true
	case SubjectOrganization:
		return ReasonOrganizationExact, 30, true
	default:
		return "", 0, false
	}
}

func sameUUID(value *uuid.UUID, expected uuid.UUID) bool {
	return value != nil && *value == expected
}

func protectedPolicyRole(fact BindingFact) bool {
	return fact.RoleID != nil && fact.RoleProtected &&
		sameUUID(fact.PolicyRoleID, *fact.RoleID) && fact.PolicyState == "active"
}

func protectedTenantAll(fact BindingFact) bool {
	return protectedPolicyRole(fact) &&
		sameUUID(fact.PolicyPermissionID, fact.PermissionID) &&
		sameUUID(fact.PolicyBindingID, fact.BindingID)
}
