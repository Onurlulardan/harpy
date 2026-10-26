package identity

import "time"

type Organization struct {
	ID           IdentityID
	Name         string
	Slug         string
	Status       OrganizationStatus
	ParentID     IdentityID
	OwnerIDs     []IdentityID
	Revision     int64
	Profile      ProfileDocument
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TombstonedAt time.Time
}

func (organization Organization) Validate() error {
	if organization.ID == (IdentityID{}) || !ValidDisplayName(organization.Name) ||
		!ValidSlug(organization.Slug) || !organization.Status.Valid() ||
		organization.ParentID == organization.ID || organization.Revision < 1 ||
		organization.OwnerIDs == nil ||
		organization.Profile.Validate(ProfileOrganization) != nil ||
		organization.CreatedAt.IsZero() ||
		organization.UpdatedAt.Before(organization.CreatedAt) ||
		(organization.Status == OrganizationTombstoned) !=
			!organization.TombstonedAt.IsZero() {
		return ErrInvalidRecord
	}
	return nil
}
