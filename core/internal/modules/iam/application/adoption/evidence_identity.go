package adoption

import (
	"sort"

	"github.com/google/uuid"
)

type membershipEvidence struct {
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
}

type assignmentEvidence struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID string    `json:"role_id"`
}

type identifierEvidence struct {
	UserID uuid.UUID `json:"user_id"`
	Kind   string    `json:"kind"`
	Value  string    `json:"value"`
}

func sortedOrganizations(snapshot Snapshot) []SourceOrganization {
	result := append([]SourceOrganization(nil), snapshot.Organizations...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID.String() < result[j].ID.String() })
	return result
}

func sortedRoles(snapshot Snapshot) []SourceRole {
	result := append([]SourceRole(nil), snapshot.Roles...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID.String() < result[j].ID.String() })
	return result
}

func sortedUsers(snapshot Snapshot) []SourceUser {
	result := append([]SourceUser(nil), snapshot.Users...)
	for index := range result {
		result[index].RoleIDs = append([]string(nil), result[index].RoleIDs...)
		sort.Strings(result[index].RoleIDs)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID.String() < result[j].ID.String() })
	return result
}

func sortedMemberships(snapshot Snapshot) []membershipEvidence {
	result := make([]membershipEvidence, 0)
	for _, user := range snapshot.Users {
		if user.OrganizationID != "" {
			result = append(result, membershipEvidence{user.ID, user.OrganizationID})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UserID.String() < result[j].UserID.String() })
	return result
}

func sortedAssignments(snapshot Snapshot) []assignmentEvidence {
	result := make([]assignmentEvidence, 0)
	for _, user := range snapshot.Users {
		for _, roleID := range user.RoleIDs {
			result = append(result, assignmentEvidence{user.ID, roleID})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i].UserID.String()+result[i].RoleID, result[j].UserID.String()+result[j].RoleID
		return left < right
	})
	return result
}
