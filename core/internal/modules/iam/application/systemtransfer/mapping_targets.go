package systemtransfer

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

func createRequiredTargets(resource publicresource.Resource) []string {
	switch resource {
	case publicresource.Users:
		return []string{publicresource.UserEmailFieldID,
			publicresource.UserRoleIDsFieldID}
	case publicresource.Roles:
		return []string{publicresource.RoleScopeKindFieldID,
			publicresource.RoleNameFieldID}
	case publicresource.Organizations:
		return []string{publicresource.OrganizationNameFieldID,
			publicresource.OrganizationSlugFieldID}
	default:
		return nil
	}
}

func protectedCoreTarget(resource publicresource.Resource, target string) bool {
	forbidden := map[publicresource.Resource]map[string]struct{}{
		publicresource.Users: setOf(publicresource.UserIDFieldID,
			publicresource.UserStatusFieldID, publicresource.UserIdentityKindFieldID,
			publicresource.UserProfileSchemaFieldID, publicresource.UserRevisionFieldID,
			publicresource.UserCreatedAtFieldID, publicresource.UserUpdatedAtFieldID),
		publicresource.Roles: setOf(publicresource.RoleIDFieldID,
			publicresource.RoleStatusFieldID, publicresource.RoleProtectedFieldID,
			publicresource.RoleDefaultFieldID, publicresource.RoleUserIDsFieldID,
			publicresource.RoleProfileSchemaFieldID, publicresource.RoleRevisionFieldID,
			publicresource.RoleCreatedAtFieldID, publicresource.RoleUpdatedAtFieldID),
		publicresource.Organizations: setOf(publicresource.OrganizationIDFieldID,
			publicresource.OrganizationStatusFieldID,
			publicresource.OrganizationUserIDsFieldID,
			publicresource.OrganizationRoleIDsFieldID,
			publicresource.OrganizationProfileFieldID,
			publicresource.OrganizationRevisionFieldID,
			publicresource.OrganizationCreatedAtFieldID,
			publicresource.OrganizationUpdatedAtFieldID),
	}
	_, found := forbidden[resource][target]
	return found
}

func setOf(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
