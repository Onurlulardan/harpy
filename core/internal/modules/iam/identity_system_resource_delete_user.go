package iam

import (
	"context"
	"fmt"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identityUserDeleteCommands interface {
	ListAssignments(
		context.Context, platformtx.TenantScope, identityfacade.ListAssignmentsQuery,
	) ([]identityfacade.Assignment, error)
	RevokeAssignment(
		context.Context, platformtx.TenantScope, identityfacade.RevokeAssignmentCommand,
	) (identityfacade.MutationResult, error)
	ListMemberships(
		context.Context, platformtx.TenantScope, identityfacade.ListMembershipsQuery,
	) ([]identityfacade.Membership, error)
	DeleteMembership(
		context.Context, platformtx.TenantScope, identityfacade.DeleteMembershipCommand,
	) (identityfacade.MutationResult, error)
	DeleteUser(
		context.Context, platformtx.TenantScope, identityfacade.UserStateCommand,
	) (identityfacade.MutationResult, error)
}

type identityUserDeleteRuntime struct {
	runtime *IdentityRuntime
}

func (runtime identityUserDeleteRuntime) ListAssignments(
	ctx context.Context, scope platformtx.TenantScope, query identityfacade.ListAssignmentsQuery,
) ([]identityfacade.Assignment, error) {
	return runtime.runtime.assignments.ListAssignments(ctx, scope, query)
}

func (runtime identityUserDeleteRuntime) RevokeAssignment(
	ctx context.Context, scope platformtx.TenantScope, command identityfacade.RevokeAssignmentCommand,
) (identityfacade.MutationResult, error) {
	return runtime.runtime.assignments.RevokeAssignment(ctx, scope, command)
}

func (runtime identityUserDeleteRuntime) ListMemberships(
	ctx context.Context, scope platformtx.TenantScope, query identityfacade.ListMembershipsQuery,
) ([]identityfacade.Membership, error) {
	return runtime.runtime.memberships.ListMemberships(ctx, scope, query)
}

func (runtime identityUserDeleteRuntime) DeleteMembership(
	ctx context.Context, scope platformtx.TenantScope, command identityfacade.DeleteMembershipCommand,
) (identityfacade.MutationResult, error) {
	return runtime.runtime.memberships.DeleteMembership(ctx, scope, command)
}

func (runtime identityUserDeleteRuntime) DeleteUser(
	ctx context.Context, scope platformtx.TenantScope, command identityfacade.UserStateCommand,
) (identityfacade.MutationResult, error) {
	return runtime.runtime.users.DeleteUser(ctx, scope, command)
}

func (port *identitySystemResourcePort) deleteSystemResourceUser(
	command identitySystemResourceCommand,
	deletion identitySystemResourceDelete,
	operationID uuid.UUID,
	requestSHA [32]byte,
	evidence identityfacade.MutationEvidence,
) (int64, bool, error) {
	if port == nil || port.application == nil ||
		port.application.identities == nil || !port.application.identities.Valid() {
		return 0, false, ErrIdentityHTTPUnavailable
	}
	return deleteIdentityUserWithRelations(
		command, deletion, operationID, requestSHA, evidence,
		identityUserDeleteRuntime{runtime: port.application.identities},
	)
}

func deleteIdentityUserWithRelations(
	command identitySystemResourceCommand,
	deletion identitySystemResourceDelete,
	operationID uuid.UUID,
	requestSHA [32]byte,
	evidence identityfacade.MutationEvidence,
	relations identityUserDeleteCommands,
) (int64, bool, error) {
	if deletion.id == command.evidence.Subject() {
		return 0, false, ErrIdentityHTTPForbidden
	}
	if relations == nil {
		return 0, false, ErrIdentityHTTPUnavailable
	}
	revision, err := terminalizeIdentityUserDeleteRelations(
		command, deletion, operationID, requestSHA, relations,
	)
	if err != nil {
		return 0, false, fmt.Errorf("delete user aggregate: %w", err)
	}
	result, err := relations.DeleteUser(
		command.ctx, command.scope, identityfacade.UserStateCommand{
			ID:               identityfacade.IdentityID(deletion.id),
			ExpectedRevision: revision,
			Evidence:         evidence,
		},
	)
	if err != nil {
		return 0, false, err
	}
	if result.ID != identityfacade.IdentityID(deletion.id) {
		return 0, false, publicresource.ErrUnavailable
	}
	return result.Revision, result.Replayed, nil
}
