package userrelation

import (
	"context"
	"reflect"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	identityports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/identity"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type TenantScope = transaction.TenantScope
type ReplaceCommand = identitydomain.ReplaceUserRelationsCommand
type OrganizationSetPatch = identitydomain.OrganizationSetPatch
type RoleSetPatch = identitydomain.RoleSetPatch
type MutationResult = identitydomain.MutationResult

const (
	CreateCommandKey = identitydomain.UserRelationCreateCommandKey
	UpdateCommandKey = identitydomain.UserRelationUpdateCommandKey
)

var (
	ErrInvalidCommand = identitydomain.ErrInvalidCommand
	ErrStorage        = identitydomain.ErrStorage
)

type Commands struct {
	store    identityports.UserRelationStore[TenantScope]
	profiles identityports.Profile[TenantScope]
}

func NewCommands(
	store identityports.UserRelationStore[TenantScope],
	profiles identityports.Profile[TenantScope],
) (*Commands, error) {
	if nilDependency(store) || nilDependency(profiles) {
		return nil, ErrStorage
	}
	return &Commands{store: store, profiles: profiles}, nil
}

func NewOrganizationSetPatch(
	userID identitydomain.IdentityID,
	replace bool,
	organizationIDs []identitydomain.IdentityID,
) (OrganizationSetPatch, error) {
	return identitydomain.NewOrganizationSetPatch(userID, replace, organizationIDs)
}

func NewRoleSetPatch(
	userID identitydomain.IdentityID,
	replace bool,
	roleIDs []identitydomain.IdentityID,
) (RoleSetPatch, error) {
	return identitydomain.NewRoleSetPatch(userID, replace, roleIDs)
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}

func mutationScopeUsable(
	ctx context.Context,
	scope TenantScope,
	evidence identitydomain.MutationEvidence,
) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) &&
		transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.Command) == nil &&
		scope.ActorID() == uuid.UUID(evidence.ActorID)
}
