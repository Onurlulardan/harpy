package iam

import (
	assignmentpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/assignment"
	bulkdeletepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/bulkdelete"
	membershippostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/membership"
	organizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/organization"
	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/profile"
	rolepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/role"
	userpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/user"
	userrelationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/userrelation"
	passwordenrollmentpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/passwordenrollment"
	assignmentapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/assignment"
	membershipapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/membership"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	userrelationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/userrelation"
	passwordenrollment "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/passwordenrollment"
	identityports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/identity"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type identityProfiles = identityports.Profile[platformtx.TenantScope]
type identityProfileAliases = identityports.ProfileAliases[platformtx.TenantScope]
type identityBulkDeleteReceipts = identityports.BulkDeleteReceiptStore[platformtx.TenantScope]

func newIdentityBulkDeleteReceipts() (identityBulkDeleteReceipts, error) {
	return bulkdeletepostgres.NewStore()
}

func newIdentityProfileStore() (identityProfileAliases, error) {
	return profilepostgres.NewStore()
}

func newIdentityAssignments() (*assignmentapp.Commands, error) {
	store, err := assignmentpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return assignmentapp.NewCommands(store)
}

func newIdentityMemberships() (*membershipapp.Commands, error) {
	store, err := membershippostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return membershipapp.NewCommands(store)
}

func newIdentityOrganizations(profiles identityProfiles) (*organizationapp.Commands, error) {
	store, err := organizationpostgres.NewStore(profiles)
	if err != nil {
		return nil, err
	}
	return organizationapp.NewCommands(store, profiles)
}

func newIdentityRoles(profiles identityProfiles) (*roleapp.Commands, error) {
	store, err := rolepostgres.NewStore(profiles)
	if err != nil {
		return nil, err
	}
	return roleapp.NewCommands(store, profiles)
}

func newIdentityUsers(profiles identityProfiles) (*userapp.Commands, error) {
	store, err := userpostgres.NewStore(profiles)
	if err != nil {
		return nil, err
	}
	return userapp.NewCommands(store, profiles)
}

func newIdentityUserRelations(profiles identityProfiles) (*userrelationapp.Commands, error) {
	store, err := userrelationpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return userrelationapp.NewCommands(store, profiles)
}

func newIdentityPasswordEnrollment() (*passwordenrollment.Commands, error) {
	store, err := passwordenrollmentpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	return passwordenrollment.NewCommands(store)
}
