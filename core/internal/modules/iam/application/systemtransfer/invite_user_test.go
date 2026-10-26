package systemtransfer

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestProcessBatchInvitesNewUserWithoutCredentialCreatePath(t *testing.T) {
	attempt := fixtureAttempt(t, publicresource.Users)
	row := fixtureSourceRow(t, publicresource.Users)
	scopes := &fixtureScopes{}
	authorizer := &fixtureAuthorizer{tenantID: fixtureTenantID,
		actorID: fixtureActorID, version: attempt.AuthorizationVersion}
	identities := &fixtureIdentityLookup{}
	resources := &fixtureResourcePort{}
	invitations := &fixtureInvitation{result: InvitationResult{
		ID: publicresource.ID(fixtureInviteID), Revision: 1,
	}}
	receipts := &fixtureReceiptStore{}
	worker, err := NewWorker(scopes, authorizer, identities, resources,
		invitations, receipts)
	if err != nil {
		t.Fatal(err)
	}
	result, err := worker.ProcessBatch(context.Background(), attempt, []SourceRow{row})
	if err != nil || result.Committed != 1 || len(invitations.inputs) != 1 {
		t.Fatalf("result=%+v invitation_calls=%d err=%v",
			result, len(invitations.inputs), err)
	}
	if resources.prepareCalls != 0 || resources.createCalls != 0 ||
		resources.updateCalls != 0 {
		t.Fatal("new user reached the password-capable system-resource create path")
	}
	wantKey, _ := RowIdempotencyKey(attempt.ID, row.Index, row.SHA256)
	if invitations.inputs[0].IdempotencyKey != wantKey ||
		invitations.inputs[0].RoleID != publicresource.ID(fixtureRoleID) {
		t.Fatalf("invitation input=%+v", invitations.inputs[0])
	}
	if len(receipts.appended) != 1 || receipts.appended[0].OwnerKind != "invitation" ||
		receipts.appended[0].OwnerID != fixtureInviteID {
		t.Fatalf("receipts=%+v", receipts.appended)
	}
}
