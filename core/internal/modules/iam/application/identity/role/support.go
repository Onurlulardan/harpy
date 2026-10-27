package role

import (
	"context"
	"reflect"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	identityports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/identity"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type TenantScope = transaction.TenantScope
type CreateRoleCommand = identitydomain.CreateRoleCommand
type GetRoleQuery = identitydomain.GetRoleQuery
type ListRolesQuery = identitydomain.ListRolesQuery
type UpdateRoleCommand = identitydomain.UpdateRoleCommand
type RetireRoleCommand = identitydomain.RetireRoleCommand
type Role = identitydomain.Role
type MutationResult = identitydomain.MutationResult
type MutationEvidence = identitydomain.MutationEvidence

const ProfileRole = identitydomain.ProfileRole

var (
	ErrInvalidCommand = identitydomain.ErrInvalidCommand
	ErrInvalidQuery   = identitydomain.ErrInvalidQuery
	ErrStorage        = identitydomain.ErrStorage
)

type Commands struct {
	store    identityports.RoleStore[TenantScope]
	profiles identityports.Profile[TenantScope]
}

func NewCommands(store identityports.RoleStore[TenantScope], profiles identityports.Profile[TenantScope]) (*Commands, error) {
	if nilInterface(store) || nilInterface(profiles) {
		return nil, ErrStorage
	}
	return &Commands{store: store, profiles: profiles}, nil
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}

func (commands *Commands) runMutation(
	ctx context.Context, scope TenantScope, evidence MutationEvidence,
	id identitydomain.IdentityID, expectedRevision int64,
	action func(context.Context, TenantScope) (MutationResult, error),
) (MutationResult, error) {
	if commands == nil || action == nil || evidence.Validate() != nil || !mutationScopeUsable(ctx, scope, evidence) {
		return MutationResult{}, ErrInvalidCommand
	}
	result, err := action(ctx, scope)
	if err != nil {
		return MutationResult{}, err
	}
	valid := result.Validate() == nil && result.ID == id && result.Revision == 1
	if expectedRevision > 0 {
		valid = result.Validate() == nil && result.ID == id && result.Revision == expectedRevision+1
	}
	if !valid {
		return MutationResult{}, ErrStorage
	}
	return result, nil
}

func mutationScopeUsable(ctx context.Context, scope TenantScope, evidence MutationEvidence) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) && transaction.ContextOwnsScope(ctx, scope) &&
		transaction.RequireActive(scope, transaction.Command) == nil && scope.ActorID() == uuid.UUID(evidence.ActorID)
}

func queryScopeUsable(ctx context.Context, scope TenantScope, evidence identitydomain.QueryEvidence) bool {
	return ctx != nil && ctx.Err() == nil && !transaction.NetworkAllowed(ctx) && transaction.ContextOwnsScope(ctx, scope) &&
		identityReadScopeActive(scope) && scope.ActorID() == uuid.UUID(evidence.ActorID)
}

func identityReadScopeActive(scope TenantScope) bool {
	return transaction.RequireActive(scope, transaction.SnapshotQuery) == nil ||
		transaction.RequireActive(scope, transaction.Command) == nil
}

func errOrStorage(err error) error {
	if err != nil {
		return err
	}
	return ErrStorage
}
