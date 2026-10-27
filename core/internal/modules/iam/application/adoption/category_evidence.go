package adoption

import "fmt"

type evidenceCategory struct {
	name  string
	count int
	value any
}

func buildCategoryEvidence(snapshot Snapshot) ([]CategoryEvidence, error) {
	categories := []evidenceCategory{
		{name: "organizations", count: len(snapshot.Organizations), value: sortedOrganizations(snapshot)},
		{name: "roles", count: len(snapshot.Roles), value: sortedRoles(snapshot)},
		{name: "users", count: len(snapshot.Users), value: sortedUsers(snapshot)},
		{name: "memberships", count: membershipCount(snapshot), value: sortedMemberships(snapshot)},
		{name: "assignments", count: assignmentCount(snapshot), value: sortedAssignments(snapshot)},
		{name: "identifiers", count: identifierCount(snapshot), value: sortedIdentifiers(snapshot)},
		{name: "portal", count: portalCount(snapshot), value: sortedPortal(snapshot)},
		{name: "credentials", count: len(snapshot.Users), value: sortedCredentials(snapshot)},
		{name: "resources", count: len(snapshot.Resources), value: sortedResources(snapshot)},
		{name: "actions", count: len(snapshot.Actions), value: sortedActions(snapshot)},
		{name: "permissions", count: len(snapshot.Permissions), value: sortedPermissions(snapshot)},
		{name: "permission_requests", count: len(snapshot.PermissionRequests), value: sortedRequests(snapshot)},
		{name: "invitations", count: len(snapshot.Invitations), value: sortedInvitations(snapshot)},
		{name: "profiles", count: len(snapshot.ProfileFields), value: sortedProfileFields(snapshot)},
		{name: "layouts", count: len(snapshot.LayoutFields), value: sortedLayoutFields(snapshot)},
	}
	result := make([]CategoryEvidence, 0, len(categories))
	for _, category := range categories {
		digest, err := canonicalSHA256(category.value)
		if err != nil {
			return nil, fmt.Errorf("canonicalize %s evidence: %w", category.name, err)
		}
		result = append(result, CategoryEvidence{
			Category: category.name, RowCount: category.count, RootSHA256: digest,
		})
	}
	return result, nil
}
