package iam

import (
	assignmentapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/assignment"
	membershipapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/membership"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	userrelationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/userrelation"
	passwordenrollment "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/passwordenrollment"
)

// IdentityRuntime owns the canonical user, role, organization and relation graph.
type IdentityRuntime struct {
	bulkDeletes   identityBulkDeleteReceipts
	assignments   *assignmentapp.Commands
	memberships   *membershipapp.Commands
	organizations *organizationapp.Commands
	roles         *roleapp.Commands
	users         *userapp.Commands
	userRelations *userrelationapp.Commands
	passwords     *passwordenrollment.Commands
	profiles      identityProfileAliases
}

func NewIdentityRuntime() (*IdentityRuntime, error) {
	profiles, err := newIdentityProfileStore()
	if err != nil {
		return nil, err
	}
	bulkDeletes, err := newIdentityBulkDeleteReceipts()
	if err != nil {
		return nil, err
	}
	assignments, err := newIdentityAssignments()
	if err != nil {
		return nil, err
	}
	memberships, err := newIdentityMemberships()
	if err != nil {
		return nil, err
	}
	organizations, err := newIdentityOrganizations(profiles)
	if err != nil {
		return nil, err
	}
	roles, err := newIdentityRoles(profiles)
	if err != nil {
		return nil, err
	}
	users, err := newIdentityUsers(profiles)
	if err != nil {
		return nil, err
	}
	userRelations, err := newIdentityUserRelations(profiles)
	if err != nil {
		return nil, err
	}
	passwords, err := newIdentityPasswordEnrollment()
	if err != nil {
		return nil, err
	}
	return &IdentityRuntime{
		assignments: assignments, memberships: memberships, bulkDeletes: bulkDeletes,
		organizations: organizations, roles: roles, users: users, passwords: passwords,
		userRelations: userRelations, profiles: profiles,
	}, nil
}

func (runtime *IdentityRuntime) Valid() bool {
	return runtime != nil && runtime.bulkDeletes != nil && runtime.assignments != nil && runtime.memberships != nil &&
		runtime.organizations != nil && runtime.roles != nil && runtime.users != nil &&
		runtime.userRelations != nil && runtime.passwords != nil && runtime.profiles != nil
}
