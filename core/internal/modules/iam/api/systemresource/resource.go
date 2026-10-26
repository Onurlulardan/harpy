package systemresource

type Resource string

const (
	Users         Resource = "users"
	Roles         Resource = "roles"
	Organizations Resource = "organizations"

	UsersCatalogResourceID         = "9c65f0c0-0000-4000-8000-000000001001"
	RolesCatalogResourceID         = "9c65f0c0-0000-4000-8000-000000001002"
	OrganizationsCatalogResourceID = "9c65f0c0-0000-4000-8000-000000001003"
)

func ParseResource(value string) (Resource, bool) {
	resource := Resource(value)
	return resource, resource.Valid()
}

func (resource Resource) CatalogResourceID() (string, error) {
	switch resource {
	case Users:
		return UsersCatalogResourceID, nil
	case Roles:
		return RolesCatalogResourceID, nil
	case Organizations:
		return OrganizationsCatalogResourceID, nil
	default:
		return "", ErrContract
	}
}

func (resource Resource) Valid() bool {
	switch resource {
	case Users, Roles, Organizations:
		return true
	default:
		return false
	}
}

func (resource Resource) AuthorityResource() (string, error) {
	if !resource.Valid() {
		return "", ErrContract
	}
	return "iam." + string(resource), nil
}
