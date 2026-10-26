package systemtransfer

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestProcessBatchReplaysOwnerCommandAfterReceiptCommitCrash(t *testing.T) {
	attempt := fixtureAttempt(t, publicresource.Users)
	row := fixtureSourceRow(t, publicresource.Users)
	scopes := &fixtureScopes{}
	authorizer := &fixtureAuthorizer{tenantID: fixtureTenantID,
		actorID: fixtureActorID, version: attempt.AuthorizationVersion}
	invitations := &fixtureInvitation{result: InvitationResult{
		ID: publicresource.ID(fixtureInviteID), Revision: 1,
	}}
	receipts := &fixtureReceiptStore{failNext: true}
	worker, _ := NewWorker(scopes, authorizer, &fixtureIdentityLookup{},
		&fixtureResourcePort{}, invitations, receipts)
	if _, err := worker.ProcessBatch(context.Background(), attempt,
		[]SourceRow{row}); err == nil {
		t.Fatal("receipt interruption was hidden")
	}
	result, err := worker.ProcessBatch(context.Background(), attempt, []SourceRow{row})
	if err != nil || result.Committed != 1 || len(invitations.inputs) != 2 ||
		scopes.calls != 2 || authorizer.calls != 2 || len(receipts.appended) != 1 {
		t.Fatalf("result=%+v scopes=%d auth=%d invites=%d receipts=%d err=%v",
			result, scopes.calls, authorizer.calls, len(invitations.inputs),
			len(receipts.appended), err)
	}
	if invitations.inputs[0].IdempotencyKey != invitations.inputs[1].IdempotencyKey {
		t.Fatal("crash replay changed the canonical owner idempotency key")
	}
}
