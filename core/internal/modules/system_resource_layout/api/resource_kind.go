package api

type ResourceKind string

const (
	ResourceUser         ResourceKind = "user"
	ResourceOrganization ResourceKind = "organization"
	ResourceRole         ResourceKind = "role"
)

func (kind ResourceKind) Valid() bool {
	return kind == ResourceUser || kind == ResourceOrganization || kind == ResourceRole
}
