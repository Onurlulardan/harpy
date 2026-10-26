package identity

const (
	organizationRelationKind = "organization_membership"
	roleRelationKind         = "role_assignment"
)

type OrganizationSetPatch struct{ set relationSetPatch }

func NewOrganizationSetPatch(
	userID IdentityID,
	replace bool,
	organizationIDs []IdentityID,
) (OrganizationSetPatch, error) {
	set, err := newRelationSetPatch(
		userID, replace, organizationIDs, organizationRelationKind,
	)
	if err != nil {
		return OrganizationSetPatch{}, err
	}
	return OrganizationSetPatch{set: set}, nil
}

func (patch OrganizationSetPatch) Replace() bool { return patch.set.replace }

func (patch OrganizationSetPatch) OrganizationIDs() []IdentityID {
	return cloneIdentityIDs(patch.set.targets)
}

func (patch OrganizationSetPatch) MembershipIDs() []IdentityID {
	return cloneIdentityIDs(patch.set.rowIDs)
}

func (patch OrganizationSetPatch) validate(userID IdentityID) error {
	return patch.set.validate(userID, organizationRelationKind)
}

type RoleSetPatch struct{ set relationSetPatch }

func NewRoleSetPatch(
	userID IdentityID,
	replace bool,
	roleIDs []IdentityID,
) (RoleSetPatch, error) {
	set, err := newRelationSetPatch(userID, replace, roleIDs, roleRelationKind)
	if err != nil {
		return RoleSetPatch{}, err
	}
	return RoleSetPatch{set: set}, nil
}

func (patch RoleSetPatch) Replace() bool { return patch.set.replace }

func (patch RoleSetPatch) RoleIDs() []IdentityID {
	return cloneIdentityIDs(patch.set.targets)
}

func (patch RoleSetPatch) AssignmentIDs() []IdentityID {
	return cloneIdentityIDs(patch.set.rowIDs)
}

func (patch RoleSetPatch) validate(userID IdentityID) error {
	return patch.set.validate(userID, roleRelationKind)
}
