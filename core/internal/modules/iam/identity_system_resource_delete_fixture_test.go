package iam

import (
	"context"
	"errors"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identityUserDeleteFake struct {
	assignmentPages   [][]identityfacade.Assignment
	membershipPages   [][]identityfacade.Membership
	assignmentPage    int
	membershipPage    int
	replayedChildren  bool
	deleteReplayed    bool
	deleteRevision    int64
	revokeErr         error
	membershipErr     error
	calls             []string
	assignmentQueries []identityfacade.ListAssignmentsQuery
	membershipQueries []identityfacade.ListMembershipsQuery
	revokes           []identityfacade.RevokeAssignmentCommand
	membershipDeletes []identityfacade.DeleteMembershipCommand
	userDeletes       []identityfacade.UserStateCommand
}

func (fake *identityUserDeleteFake) ListAssignments(
	_ context.Context, _ platformtx.TenantScope, query identityfacade.ListAssignmentsQuery,
) ([]identityfacade.Assignment, error) {
	fake.calls = append(fake.calls, "list-assignments")
	fake.assignmentQueries = append(fake.assignmentQueries, query)
	if fake.assignmentPage >= len(fake.assignmentPages) {
		return nil, nil
	}
	result := fake.assignmentPages[fake.assignmentPage]
	fake.assignmentPage++
	return result, nil
}

func (fake *identityUserDeleteFake) RevokeAssignment(
	_ context.Context, _ platformtx.TenantScope, command identityfacade.RevokeAssignmentCommand,
) (identityfacade.MutationResult, error) {
	fake.calls = append(fake.calls, "revoke-assignment")
	fake.revokes = append(fake.revokes, command)
	if fake.revokeErr != nil {
		return identityfacade.MutationResult{}, fake.revokeErr
	}
	return identityfacade.MutationResult{
		ID: command.ID, Revision: command.ExpectedRevision + 1,
		AuthorizationClock: 1, Replayed: fake.replayedChildren,
	}, nil
}

func (fake *identityUserDeleteFake) ListMemberships(
	_ context.Context, _ platformtx.TenantScope, query identityfacade.ListMembershipsQuery,
) ([]identityfacade.Membership, error) {
	fake.calls = append(fake.calls, "list-memberships")
	fake.membershipQueries = append(fake.membershipQueries, query)
	if fake.membershipPage >= len(fake.membershipPages) {
		return nil, nil
	}
	result := fake.membershipPages[fake.membershipPage]
	fake.membershipPage++
	return result, nil
}

func (fake *identityUserDeleteFake) DeleteMembership(
	_ context.Context, _ platformtx.TenantScope, command identityfacade.DeleteMembershipCommand,
) (identityfacade.MutationResult, error) {
	fake.calls = append(fake.calls, "delete-membership")
	fake.membershipDeletes = append(fake.membershipDeletes, command)
	if fake.membershipErr != nil {
		return identityfacade.MutationResult{}, fake.membershipErr
	}
	return identityfacade.MutationResult{
		ID: command.ID, Revision: command.ExpectedRevision + 1,
		AuthorizationClock: 1, Replayed: fake.replayedChildren,
	}, nil
}

func (fake *identityUserDeleteFake) DeleteUser(
	_ context.Context, _ platformtx.TenantScope, command identityfacade.UserStateCommand,
) (identityfacade.MutationResult, error) {
	fake.calls = append(fake.calls, "delete-user")
	fake.userDeletes = append(fake.userDeletes, command)
	revision := fake.deleteRevision
	if revision == 0 {
		revision = command.ExpectedRevision + 1
	}
	return identityfacade.MutationResult{
		ID: command.ID, Revision: revision,
		AuthorizationClock: 1, Replayed: fake.deleteReplayed,
	}, nil
}

var errIdentityUserDeleteChild = errors.New("identity user delete child failed")

func identityUserDeleteTestCommand() (
	identitySystemResourceCommand,
	identitySystemResourceDelete,
	uuid.UUID,
	[32]byte,
	identityfacade.MutationEvidence,
) {
	actorID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	tenantID := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	permissionID := uuid.MustParse("10000000-0000-4000-8000-000000000003")
	operationID := uuid.MustParse("10000000-0000-4000-8000-000000000004")
	targetID := uuid.MustParse("10000000-0000-4000-8000-000000000005")
	requestSHA := [32]byte{1, 2, 3}
	publicEvidence := identityPublicEvidence{
		tenantID: tenantID, actorID: actorID,
		tokenVersion: 3, authorizationVersion: 5,
	}
	command := identitySystemResourceCommand{
		ctx: context.Background(), evidence: publicEvidence,
		permissionID: permissionID,
	}
	deletion := identitySystemResourceDelete{id: targetID, expectedRevision: 7}
	evidence := identityMutationEvidence(
		publicEvidence, permissionID, operationID, requestSHA,
	)
	return command, deletion, operationID, requestSHA, evidence
}
