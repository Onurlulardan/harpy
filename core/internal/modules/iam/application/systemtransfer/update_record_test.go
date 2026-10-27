package systemtransfer

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestUpdateRecordUsesObservedRevisionCAS(t *testing.T) {
	attempt := fixtureAttempt(t, publicresource.Roles)
	row := fixtureSourceRow(t, publicresource.Roles)
	mutation, err := fixtureMutationResult(publicresource.Roles,
		publicresource.ID(fixtureRecordID), 5, false)
	if err != nil {
		t.Fatal(err)
	}
	resources := &fixtureResourcePort{result: mutation}
	worker, _ := NewWorker(&fixtureScopes{}, &fixtureAuthorizer{
		tenantID: fixtureTenantID, actorID: fixtureActorID,
		version: attempt.AuthorizationVersion,
	}, &fixtureIdentityLookup{match: fixtureMatch(true, 4)}, resources,
		&fixtureInvitation{}, &fixtureReceiptStore{})
	result, err := worker.ProcessBatch(context.Background(), attempt, []SourceRow{row})
	wantKey, _ := RowIdempotencyKey(attempt.ID, row.Index, row.SHA256)
	if err != nil || result.Committed != 1 || resources.updateCalls != 1 ||
		resources.updateRevision != 4 || resources.updateKey != wantKey {
		t.Fatalf("result=%+v port=%+v err=%v", result, resources, err)
	}
}
