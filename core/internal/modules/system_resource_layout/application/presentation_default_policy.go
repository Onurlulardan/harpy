package application

import layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"

const (
	userEmail            = "a1000000-0000-4000-8000-000000000002"
	userPortalIdentifier = "a1000000-0000-4000-8000-000000000003"
	userStatus           = "a1000000-0000-4000-8000-000000000004"
	userIdentityKind     = "a1000000-0000-4000-8000-000000000005"
	userRoleIDs          = "a1000000-0000-4000-8000-000000000006"
	userOrganizationIDs  = "a1000000-0000-4000-8000-000000000007"
	userCreatedAt        = "a1000000-0000-4000-8000-00000000000a"
	userUpdatedAt        = "a1000000-0000-4000-8000-00000000000b"
	roleScopeKind        = "b1000000-0000-4000-8000-000000000002"
	roleOrganizationID   = "b1000000-0000-4000-8000-000000000004"
	roleName             = "b1000000-0000-4000-8000-000000000005"
	roleDescription      = "b1000000-0000-4000-8000-000000000006"
	roleStatus           = "b1000000-0000-4000-8000-000000000007"
	roleProtected        = "b1000000-0000-4000-8000-000000000008"
	roleDefault          = "b1000000-0000-4000-8000-000000000009"
	roleCreatedAt        = "b1000000-0000-4000-8000-00000000000d"
	roleUpdatedAt        = "b1000000-0000-4000-8000-00000000000e"
	organizationName     = "c1000000-0000-4000-8000-000000000002"
	organizationSlug     = "c1000000-0000-4000-8000-000000000003"
	organizationStatus   = "c1000000-0000-4000-8000-000000000004"
	organizationParentID = "c1000000-0000-4000-8000-000000000005"
	organizationCreated  = "c1000000-0000-4000-8000-00000000000a"
	organizationUpdated  = "c1000000-0000-4000-8000-00000000000b"
	organizationOwnerIDs = "c1000000-0000-4000-8000-00000000000c"
)

type defaultSurfacePolicy struct {
	ids           []string
	includeCustom bool
}

func defaultSurfaceFieldIDs(
	resource layoutapi.ResourceKind,
	surface layoutapi.SurfaceKind,
) ([]string, bool, bool) {
	policy, found := defaultSurfacePolicies[resource][surface]
	return policy.ids, policy.includeCustom, found
}

var defaultSurfacePolicies = map[layoutapi.ResourceKind]map[layoutapi.SurfaceKind]defaultSurfacePolicy{
	layoutapi.ResourceUser: {
		layoutapi.SurfaceRecordCreate: {[]string{userEmail, userOrganizationIDs, userRoleIDs}, true},
		layoutapi.SurfaceRecordUpdate: {[]string{userOrganizationIDs, userRoleIDs}, true},
		layoutapi.SurfaceRecordDetail: {[]string{userEmail, userStatus, userIdentityKind, userOrganizationIDs, userRoleIDs, userCreatedAt, userUpdatedAt}, true},
		layoutapi.SurfaceRecordList:   {[]string{userEmail, userStatus, userIdentityKind, userOrganizationIDs, userRoleIDs, userCreatedAt}, true},
		layoutapi.SurfaceRecordFilter: {[]string{userEmail, userStatus, userIdentityKind}, true},
	},
	layoutapi.ResourceRole: {
		layoutapi.SurfaceRecordCreate: {[]string{roleName, roleDescription, roleScopeKind, roleOrganizationID, roleDefault}, false},
		layoutapi.SurfaceRecordUpdate: {[]string{roleName, roleDescription, roleDefault}, false},
		layoutapi.SurfaceRecordDetail: {[]string{roleName, roleDescription, roleStatus, roleScopeKind, roleOrganizationID, roleProtected, roleDefault, roleCreatedAt, roleUpdatedAt}, false},
		layoutapi.SurfaceRecordList:   {[]string{roleName, roleDescription, roleStatus, roleScopeKind, roleOrganizationID, roleProtected, roleDefault, roleCreatedAt}, false},
		layoutapi.SurfaceRecordFilter: {[]string{roleName, roleStatus, roleScopeKind, roleProtected, roleDefault}, false},
	},
	layoutapi.ResourceOrganization: {
		layoutapi.SurfaceRecordCreate: {[]string{organizationName, organizationSlug, organizationParentID, organizationOwnerIDs}, false},
		layoutapi.SurfaceRecordUpdate: {[]string{organizationName, organizationSlug, organizationParentID, organizationOwnerIDs}, false},
		layoutapi.SurfaceRecordDetail: {[]string{organizationName, organizationSlug, organizationStatus, organizationParentID, organizationOwnerIDs, organizationCreated, organizationUpdated}, false},
		layoutapi.SurfaceRecordList:   {[]string{organizationName, organizationSlug, organizationStatus, organizationParentID, organizationOwnerIDs, organizationCreated}, false},
		layoutapi.SurfaceRecordFilter: {[]string{organizationName, organizationSlug, organizationStatus}, false},
	},
}

func defaultCoreFieldID(resource layoutapi.ResourceKind, id string) bool {
	prefix := map[layoutapi.ResourceKind]string{
		layoutapi.ResourceUser: "a1000000-", layoutapi.ResourceRole: "b1000000-",
		layoutapi.ResourceOrganization: "c1000000-",
	}[resource]
	return len(prefix) > 0 && len(id) >= len(prefix) && id[:len(prefix)] == prefix
}
