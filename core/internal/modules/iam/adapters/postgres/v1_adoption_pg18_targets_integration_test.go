//go:build integration

package postgres

func adoptedTargets(
	tenant iamFoundationTenant, category string,
) []iamV1AdoptionTarget {
	switch category {
	case "organizations":
		return adoptedOrganizationTargets(tenant)
	case "roles":
		return adoptedRoleTargets(tenant)
	case "users":
		return adoptedUserTargets(tenant)
	case "memberships":
		return adoptedMembershipTargets(tenant)
	case "assignments":
		return adoptedAssignmentTargets(tenant)
	case "identifiers":
		return adoptedIdentifierTargets(tenant)
	case "portal":
		return adoptedPortalTargets(tenant)
	case "credentials":
		return adoptedCredentialTargets(tenant)
	case "resources":
		return adoptedResourceTargets(tenant)
	case "actions":
		return adoptedActionTargets(tenant)
	case "permissions":
		return adoptedPermissionTargets(tenant)
	case "permission_requests":
		return adoptedPermissionRequestTargets(tenant)
	case "invitations":
		return adoptedInvitationTargets(tenant)
	case "profiles":
		return adoptedProfileTargets(tenant)
	default:
		return nil
	}
}
