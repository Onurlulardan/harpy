package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type UserStore[Scope any] interface {
	CreateUser(context.Context, Scope, identitydomain.CreateUserCommand) (identitydomain.MutationResult, error)
	GetUser(context.Context, Scope, identitydomain.GetUserQuery) (identitydomain.User, error)
	ListUsers(context.Context, Scope, identitydomain.ListUsersQuery) ([]identitydomain.User, error)
	UpdateUser(context.Context, Scope, identitydomain.UpdateUserCommand) (identitydomain.MutationResult, error)
	DisableUser(context.Context, Scope, identitydomain.UserStateCommand) (identitydomain.MutationResult, error)
	DeleteUser(context.Context, Scope, identitydomain.UserStateCommand) (identitydomain.MutationResult, error)
}
