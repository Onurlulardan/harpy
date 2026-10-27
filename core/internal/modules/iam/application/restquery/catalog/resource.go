package catalog

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const Owner = "iam"

type Resource struct {
	kind           profile.ResourceKind
	resourceID     uuid.UUID
	resourceKey    string
	table          string
	statusFieldID  uuid.UUID
	terminalStatus string
}

var resources = map[profile.ResourceKind]Resource{
	profile.ResourceUser: {
		kind:        profile.ResourceUser,
		resourceID:  uuid.MustParse(publicresource.UsersCatalogResourceID),
		resourceKey: "iam.users", table: "iam_rest_query_users_v2",
		statusFieldID: uuid.MustParse(userStatusFieldID), terminalStatus: "tombstoned",
	},
	profile.ResourceRole: {
		kind:        profile.ResourceRole,
		resourceID:  uuid.MustParse(publicresource.RolesCatalogResourceID),
		resourceKey: "iam.roles", table: "iam_rest_query_roles_v2",
		statusFieldID: uuid.MustParse(roleStatusFieldID), terminalStatus: "retired",
	},
	profile.ResourceOrganization: {
		kind:        profile.ResourceOrganization,
		resourceID:  uuid.MustParse(publicresource.OrganizationsCatalogResourceID),
		resourceKey: "iam.organizations", table: "iam_rest_query_organizations_v3",
		statusFieldID: uuid.MustParse(organizationStatusFieldID), terminalStatus: "tombstoned",
	},
}

func ResourceFor(kind profile.ResourceKind) (Resource, error) {
	resource, found := resources[kind]
	if !found {
		return Resource{}, ErrContract
	}
	return resource, nil
}

func ResourceByID(id uuid.UUID) (Resource, error) {
	for _, resource := range resources {
		if resource.resourceID == id {
			return resource, nil
		}
	}
	return Resource{}, ErrContract
}

func (resource Resource) Kind() profile.ResourceKind { return resource.kind }
func (resource Resource) ID() uuid.UUID              { return resource.resourceID }
func (resource Resource) Key() string                { return resource.resourceKey }
func (resource Resource) Table() string              { return resource.table }
func (resource Resource) StatusFieldID() uuid.UUID   { return resource.statusFieldID }
func (resource Resource) TerminalStatus() string     { return resource.terminalStatus }
