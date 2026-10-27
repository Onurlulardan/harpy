package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type OrganizationStore[Scope any] interface {
	CreateOrganization(context.Context, Scope, identitydomain.CreateOrganizationCommand) (identitydomain.MutationResult, error)
	GetOrganization(context.Context, Scope, identitydomain.GetOrganizationQuery) (identitydomain.Organization, error)
	ListOrganizations(context.Context, Scope, identitydomain.ListOrganizationsQuery) ([]identitydomain.Organization, error)
	UpdateOrganization(context.Context, Scope, identitydomain.UpdateOrganizationCommand) (identitydomain.MutationResult, error)
	DisableOrganization(context.Context, Scope, identitydomain.OrganizationStateCommand) (identitydomain.MutationResult, error)
	DeleteOrganization(context.Context, Scope, identitydomain.OrganizationStateCommand) (identitydomain.MutationResult, error)
}
