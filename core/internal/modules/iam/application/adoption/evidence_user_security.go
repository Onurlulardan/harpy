package adoption

import "sort"

func sortedIdentifiers(snapshot Snapshot) []identifierEvidence {
	result := make([]identifierEvidence, 0)
	for _, user := range snapshot.Users {
		if user.Email != "" {
			result = append(result, identifierEvidence{user.ID, "email_v1", user.Email})
		}
		_, portal := sourceAccessModes(user.AccountType)
		if portal && user.PortalLoginIdentifier != "" {
			result = append(result, identifierEvidence{user.ID, "portal_v1", user.PortalLoginIdentifier})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i].UserID.String() + result[i].Kind
		right := result[j].UserID.String() + result[j].Kind
		return left < right
	})
	return result
}

func sortedPortal(snapshot Snapshot) []SourceUser {
	result := make([]SourceUser, 0)
	for _, user := range sortedUsers(snapshot) {
		if _, portal := sourceAccessModes(user.AccountType); portal {
			result = append(result, user)
		}
	}
	return result
}

type credentialEvidence struct {
	UserID       string `json:"user_id"`
	PasswordHash string `json:"password_hash"`
	TokenVersion int64  `json:"token_version"`
}

func sortedCredentials(snapshot Snapshot) []credentialEvidence {
	users := sortedUsers(snapshot)
	result := make([]credentialEvidence, 0, len(users))
	for _, user := range users {
		result = append(result, credentialEvidence{
			UserID: user.ID.String(), PasswordHash: user.PasswordHash,
			TokenVersion: user.TokenVersion,
		})
	}
	return result
}

func membershipCount(snapshot Snapshot) int {
	return len(sortedMemberships(snapshot))
}

func assignmentCount(snapshot Snapshot) int {
	return len(sortedAssignments(snapshot))
}

func identifierCount(snapshot Snapshot) int {
	return len(sortedIdentifiers(snapshot))
}

func portalCount(snapshot Snapshot) int {
	return len(sortedPortal(snapshot))
}
