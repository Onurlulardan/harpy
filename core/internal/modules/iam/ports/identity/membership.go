package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type MembershipStore[Scope any] interface {
	AddMembership(context.Context, Scope, identitydomain.AddMembershipCommand) (identitydomain.MutationResult, error)
	GetMembership(context.Context, Scope, identitydomain.GetMembershipQuery) (identitydomain.Membership, error)
	ListMemberships(context.Context, Scope, identitydomain.ListMembershipsQuery) ([]identitydomain.Membership, error)
	UpdateMembership(context.Context, Scope, identitydomain.UpdateMembershipCommand) (identitydomain.MutationResult, error)
	DeleteMembership(context.Context, Scope, identitydomain.DeleteMembershipCommand) (identitydomain.MutationResult, error)
}
