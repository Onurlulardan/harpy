package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type RoleStore[Scope any] interface {
	CreateRole(context.Context, Scope, identitydomain.CreateRoleCommand) (identitydomain.MutationResult, error)
	GetRole(context.Context, Scope, identitydomain.GetRoleQuery) (identitydomain.Role, error)
	ListRoles(context.Context, Scope, identitydomain.ListRolesQuery) ([]identitydomain.Role, error)
	UpdateRole(context.Context, Scope, identitydomain.UpdateRoleCommand) (identitydomain.MutationResult, error)
	RetireRole(context.Context, Scope, identitydomain.RetireRoleCommand) (identitydomain.MutationResult, error)
}
