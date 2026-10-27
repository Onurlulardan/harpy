package iam

import (
	"errors"
	"slices"
	"testing"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func TestDeleteIdentityUserTerminalizesOnlyTargetRelationsInOrder(t *testing.T) {
	command, deletion, operationID, requestSHA, evidence := identityUserDeleteTestCommand()
	otherUser := identityfacade.IdentityID(
		uuid.MustParse("20000000-0000-4000-8000-000000000001"),
	)
	assignmentID := identityfacade.IdentityID(
		uuid.MustParse("20000000-0000-4000-8000-000000000002"),
	)
	membershipID := identityfacade.IdentityID(
		uuid.MustParse("20000000-0000-4000-8000-000000000003"),
	)
	targetID := identityfacade.IdentityID(deletion.id)
	fake := &identityUserDeleteFake{
		assignmentPages: [][]identityfacade.Assignment{{
			{ID: assignmentID, UserID: targetID, State: identityfacade.AssignmentSuspended, Revision: 2},
			{ID: identityfacade.IdentityID(uuid.New()), UserID: otherUser, State: identityfacade.AssignmentActive, Revision: 4},
			{ID: identityfacade.IdentityID(uuid.New()), UserID: targetID, State: identityfacade.AssignmentRevoked, Revision: 6},
		}},
		membershipPages: [][]identityfacade.Membership{{
			{ID: membershipID, UserID: targetID, State: identityfacade.MembershipSuspended, Revision: 3},
			{ID: identityfacade.IdentityID(uuid.New()), UserID: otherUser, State: identityfacade.MembershipActive, Revision: 2},
			{ID: identityfacade.IdentityID(uuid.New()), UserID: targetID, State: identityfacade.MembershipTombstoned, Revision: 8},
		}},
	}
	revision, replayed, err := deleteIdentityUserWithRelations(
		command, deletion, operationID, requestSHA, evidence, fake,
	)
	if err != nil || replayed || revision != 10 {
		t.Fatalf("delete user result = %d/%t: %v", revision, replayed, err)
	}
	wantCalls := []string{
		"list-assignments", "revoke-assignment", "list-memberships",
		"delete-membership", "delete-user",
	}
	if !slices.Equal(fake.calls, wantCalls) {
		t.Fatalf("delete calls = %v, want %v", fake.calls, wantCalls)
	}
	if len(fake.revokes) != 1 || fake.revokes[0].ID != assignmentID ||
		len(fake.membershipDeletes) != 1 || fake.membershipDeletes[0].ID != membershipID ||
		len(fake.userDeletes) != 1 || fake.userDeletes[0].ExpectedRevision != 9 {
		t.Fatalf("unexpected delete graph: %+v/%+v/%+v",
			fake.revokes, fake.membershipDeletes, fake.userDeletes)
	}
	assertIdentityUserDeleteChildEvidence(
		t, command, operationID, requestSHA,
		"assignment-revoke", assignmentID, fake.revokes[0].Evidence,
	)
	assertIdentityUserDeleteChildEvidence(
		t, command, operationID, requestSHA,
		"membership-delete", membershipID, fake.membershipDeletes[0].Evidence,
	)
	for _, limit := range []int{
		fake.assignmentQueries[0].Limit, fake.membershipQueries[0].Limit,
	} {
		if limit != identityUserDeleteRelationBatchSize {
			t.Fatalf("relation query limit = %d", limit)
		}
	}
}

func TestDeleteIdentityUserReplaySkipsTerminalChildrenAndReturnsStableReceipt(t *testing.T) {
	command, deletion, operationID, requestSHA, evidence := identityUserDeleteTestCommand()
	targetID := identityfacade.IdentityID(deletion.id)
	fake := &identityUserDeleteFake{
		assignmentPages: [][]identityfacade.Assignment{{
			{ID: identityfacade.IdentityID(uuid.New()), UserID: targetID, State: identityfacade.AssignmentRevoked, Revision: 2},
		}},
		membershipPages: [][]identityfacade.Membership{{
			{ID: identityfacade.IdentityID(uuid.New()), UserID: targetID, State: identityfacade.MembershipTombstoned, Revision: 2},
		}},
		deleteReplayed: true, deleteRevision: 10,
	}
	revision, replayed, err := deleteIdentityUserWithRelations(
		command, deletion, operationID, requestSHA, evidence, fake,
	)
	if err != nil || !replayed || revision != 10 {
		t.Fatalf("replay result = %d/%t: %v", revision, replayed, err)
	}
	if len(fake.revokes) != 0 || len(fake.membershipDeletes) != 0 ||
		len(fake.userDeletes) != 1 || fake.userDeletes[0].ExpectedRevision != 7 {
		t.Fatalf("replay mutated terminal children: %+v", fake.calls)
	}
}

func TestDeleteIdentityUserStopsOnChildFailureAndRejectsSelf(t *testing.T) {
	command, deletion, operationID, requestSHA, evidence := identityUserDeleteTestCommand()
	targetID := identityfacade.IdentityID(deletion.id)
	fake := &identityUserDeleteFake{
		assignmentPages: [][]identityfacade.Assignment{{
			{ID: identityfacade.IdentityID(uuid.New()), UserID: targetID, State: identityfacade.AssignmentActive, Revision: 1},
		}},
		revokeErr: errIdentityUserDeleteChild,
	}
	_, _, err := deleteIdentityUserWithRelations(
		command, deletion, operationID, requestSHA, evidence, fake,
	)
	if !errors.Is(err, errIdentityUserDeleteChild) ||
		len(fake.membershipQueries) != 0 || len(fake.userDeletes) != 0 {
		t.Fatalf("child failure result = %v calls=%v", err, fake.calls)
	}
	deletion.id = command.evidence.Subject()
	self := &identityUserDeleteFake{}
	_, _, err = deleteIdentityUserWithRelations(
		command, deletion, operationID, requestSHA, evidence, self,
	)
	if !errors.Is(err, ErrIdentityHTTPForbidden) || len(self.calls) != 0 {
		t.Fatalf("self delete result = %v calls=%v", err, self.calls)
	}
}

func assertIdentityUserDeleteChildEvidence(
	t *testing.T,
	command identitySystemResourceCommand,
	operationID uuid.UUID,
	requestSHA [32]byte,
	purpose string,
	relationID identityfacade.IdentityID,
	actual identityfacade.MutationEvidence,
) {
	t.Helper()
	expectedID := identityChildID(
		operationID, purpose+":"+uuid.UUID(relationID).String(),
	)
	if actual.OperationID != identityfacade.IdentityID(expectedID) ||
		actual.RequestSHA256 != requestSHA ||
		actual.ActorID != identityfacade.IdentityID(command.evidence.Subject()) ||
		actual.PermissionID != identityfacade.IdentityID(command.permissionID) {
		t.Fatalf("child evidence = %+v", actual)
	}
}
