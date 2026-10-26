package systemtransfer

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestAuthorizeBatchRequiresFreshPinnedAuthorizationCheckpoint(t *testing.T) {
	attempt := fixtureAttempt(t, publicresource.Users)
	authorizer := &fixtureAuthorizer{tenantID: fixtureTenantID,
		actorID: fixtureActorID, version: attempt.AuthorizationVersion + 1}
	invitations := &fixtureInvitation{}
	worker, _ := NewWorker(&fixtureScopes{}, authorizer, &fixtureIdentityLookup{},
		&fixtureResourcePort{}, invitations, &fixtureReceiptStore{})
	_, err := worker.ProcessBatch(context.Background(), attempt,
		[]SourceRow{fixtureSourceRow(t, publicresource.Users)})
	if err != ErrAuthorizationCheckpoint || authorizer.calls != 1 ||
		len(invitations.inputs) != 0 {
		t.Fatalf("auth_calls=%d invitation_calls=%d err=%v",
			authorizer.calls, len(invitations.inputs), err)
	}
}
