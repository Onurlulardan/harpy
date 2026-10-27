package identity

import "time"

type User struct {
	ID                   IdentityID
	Status               UserStatus
	TokenVersion         int64
	AuthorizationVersion int64
	Revision             int64
	IdentifierID         IdentityID
	EmailDisplay         string
	EmailNormalized      string
	RoleIDs              []IdentityID
	OrganizationIDs      []IdentityID
	EmailVerified        bool
	AccessModes          []string
	Profile              ProfileDocument
	CreatedAt            time.Time
	UpdatedAt            time.Time
	TombstonedAt         time.Time
}

func (user User) Validate() error {
	if user.ID == (IdentityID{}) || !user.Status.Valid() || user.TokenVersion < 1 ||
		user.AuthorizationVersion < 0 || user.Revision < 1 ||
		user.IdentifierID == (IdentityID{}) || user.EmailDisplay == "" ||
		user.EmailNormalized == "" || user.RoleIDs == nil ||
		user.OrganizationIDs == nil ||
		!validUserAccessModes(user.AccessModes) ||
		user.Profile.Validate(ProfileUser) != nil || user.CreatedAt.IsZero() ||
		user.UpdatedAt.Before(user.CreatedAt) ||
		(user.Status == UserTombstoned) != !user.TombstonedAt.IsZero() {
		return ErrInvalidRecord
	}
	return nil
}

func validUserAccessModes(values []string) bool {
	if values == nil || len(values) > 2 {
		return false
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "operator_console" && value != "portal" {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
