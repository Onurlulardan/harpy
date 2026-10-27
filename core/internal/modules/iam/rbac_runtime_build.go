package iam

import (
	actionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/action"
	assignmentpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/assignment"
	implicationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/implication"
	organizationbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/organizationbinding"
	permissionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/permission"
	protectedadminpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/protectedadmin"
	resourcepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/resource"
	rolebindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/rolebinding"
	userbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbac/userbinding"
	actionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/action"
	assignmentapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/assignment"
	implicationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/implication"
	organizationbindingapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/organizationbinding"
	permissionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/permission"
	protectedadminapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/protectedadmin"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/resource"
	rolebindingapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/rolebinding"
	userbindingapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/userbinding"
	rbacports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type rbacSink = rbacports.ChangeSink[platformtx.TenantScope]

func buildRBACActions(sink rbacSink) (*actionapp.Commands, error) {
	store, err := actionpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return actionapp.NewCommands(store, sink)
}

func buildRBACAssignments(sink rbacSink) (*assignmentapp.Commands, error) {
	store, err := assignmentpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return assignmentapp.NewCommands(store, sink)
}

func buildRBACImplications(sink rbacSink) (*implicationapp.Commands, error) {
	store, err := implicationpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return implicationapp.NewCommands(store, sink)
}

func buildRBACOrganizationBindings(sink rbacSink) (*organizationbindingapp.Commands, error) {
	store, err := organizationbindingpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return organizationbindingapp.NewCommands(store, sink)
}

func buildRBACPermissions(sink rbacSink) (*permissionapp.Commands, error) {
	store, err := permissionpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return permissionapp.NewCommands(store, sink)
}

func buildRBACProtectedAdmins(sink rbacSink) (*protectedadminapp.Commands, error) {
	store, err := protectedadminpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return protectedadminapp.NewCommands(store, sink)
}

func buildRBACResources(sink rbacSink) (*resourceapp.Commands, error) {
	store, err := resourcepostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return resourceapp.NewCommands(store, sink)
}

func buildRBACRoleBindings(sink rbacSink) (*rolebindingapp.Commands, error) {
	store, err := rolebindingpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return rolebindingapp.NewCommands(store, sink)
}

func buildRBACUserBindings(sink rbacSink) (*userbindingapp.Commands, error) {
	store, err := userbindingpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return userbindingapp.NewCommands(store, sink)
}
