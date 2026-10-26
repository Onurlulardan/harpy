package assignment

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	identityports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/identity"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type TenantScope = transaction.TenantScope
type GrantAssignmentCommand = identitydomain.GrantAssignmentCommand
type GetAssignmentQuery = identitydomain.GetAssignmentQuery
type ListAssignmentsQuery = identitydomain.ListAssignmentsQuery
type UpdateAssignmentCommand = identitydomain.UpdateAssignmentCommand
type RevokeAssignmentCommand = identitydomain.RevokeAssignmentCommand
type Assignment = identitydomain.Assignment
type MutationResult = identitydomain.MutationResult
type MutationEvidence = identitydomain.MutationEvidence

var (
	ErrInvalidCommand = identitydomain.ErrInvalidCommand
	ErrInvalidQuery   = identitydomain.ErrInvalidQuery
	ErrStorage        = identitydomain.ErrStorage
)

type Commands struct {
	store identityports.AssignmentStore[TenantScope]
}

func NewCommands(store identityports.AssignmentStore[TenantScope]) (*Commands, error) {
	if store == nil {
		return nil, ErrStorage
	}
	return &Commands{store: store}, nil
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
