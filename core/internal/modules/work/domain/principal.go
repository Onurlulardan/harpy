package domain

type PrincipalKind string

const (
	PrincipalUser         PrincipalKind = "USER"
	PrincipalRole         PrincipalKind = "ROLE"
	PrincipalOrganization PrincipalKind = "ORGANIZATION"
)

func (kind PrincipalKind) Valid() bool {
	return kind == PrincipalUser || kind == PrincipalRole || kind == PrincipalOrganization
}

type Principal struct {
	Kind PrincipalKind
	ID   ID
}

func (principal Principal) Valid() bool {
	return principal.Kind.Valid() && principal.ID.Valid()
}

type BoardRole string

const (
	RoleViewer BoardRole = "VIEWER"
	RoleEditor BoardRole = "EDITOR"
	RoleAdmin  BoardRole = "ADMIN"
)

func (role BoardRole) Valid() bool {
	return role == RoleViewer || role == RoleEditor || role == RoleAdmin
}

func (role BoardRole) Rank() int {
	switch role {
	case RoleViewer:
		return 1
	case RoleEditor:
		return 2
	case RoleAdmin:
		return 3
	default:
		return 0
	}
}

func HighestRole(roles ...BoardRole) BoardRole {
	result := BoardRole("")
	for _, role := range roles {
		if role.Rank() > result.Rank() {
			result = role
		}
	}
	return result
}
