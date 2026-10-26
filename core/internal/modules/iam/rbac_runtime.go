package iam

import (
	actionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/action"
	assignmentapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/assignment"
	implicationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/implication"
	organizationbindingapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/organizationbinding"
	permissionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/permission"
	protectedadminapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/protectedadmin"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/resource"
	rolebindingapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/rolebinding"
	userbindingapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/userbinding"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

// RBACRuntime is the module-owned, transaction-bound RBAC application graph.
// Its application services stay private so composition cannot bypass IAM.
type RBACRuntime struct {
	actions              *actionapp.Commands
	assignments          *assignmentapp.Commands
	implications         *implicationapp.Commands
	organizationBindings *organizationbindingapp.Commands
	permissions          *permissionapp.Commands
	protectedAdmins      *protectedadminapp.Commands
	resources            *resourceapp.Commands
	roleBindings         *rolebindingapp.Commands
	userBindings         *userbindingapp.Commands
}

func NewRBACRuntime(catalogs platformasync.WriteCatalogs, now RBACClock) (*RBACRuntime, error) {
	sink, err := NewRBACChangeSink(catalogs, now)
	if err != nil {
		return nil, err
	}
	actions, err := buildRBACActions(sink)
	if err != nil {
		return nil, err
	}
	assignments, err := buildRBACAssignments(sink)
	if err != nil {
		return nil, err
	}
	implications, err := buildRBACImplications(sink)
	if err != nil {
		return nil, err
	}
	organizationBindings, err := buildRBACOrganizationBindings(sink)
	if err != nil {
		return nil, err
	}
	permissions, err := buildRBACPermissions(sink)
	if err != nil {
		return nil, err
	}
	protectedAdmins, err := buildRBACProtectedAdmins(sink)
	if err != nil {
		return nil, err
	}
	resources, err := buildRBACResources(sink)
	if err != nil {
		return nil, err
	}
	roleBindings, err := buildRBACRoleBindings(sink)
	if err != nil {
		return nil, err
	}
	userBindings, err := buildRBACUserBindings(sink)
	if err != nil {
		return nil, err
	}
	return &RBACRuntime{
		actions: actions, assignments: assignments, implications: implications,
		organizationBindings: organizationBindings, permissions: permissions,
		protectedAdmins: protectedAdmins, resources: resources,
		roleBindings: roleBindings, userBindings: userBindings,
	}, nil
}

func (runtime *RBACRuntime) Valid() bool {
	return runtime != nil && runtime.actions != nil && runtime.assignments != nil &&
		runtime.implications != nil &&
		runtime.organizationBindings != nil && runtime.permissions != nil &&
		runtime.protectedAdmins != nil && runtime.resources != nil &&
		runtime.roleBindings != nil && runtime.userBindings != nil
}
