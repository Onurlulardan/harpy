package systemtransfer

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestProcessBatchDuplicateSkipDoesNotMutate(t *testing.T) {
	attempt := fixtureAttempt(t, publicresource.Roles)
	attempt.DuplicateAction = DuplicateSkip
	resources := &fixtureResourcePort{}
	receipts := &fixtureReceiptStore{}
	worker, _ := NewWorker(&fixtureScopes{}, &fixtureAuthorizer{
		tenantID: fixtureTenantID, actorID: fixtureActorID,
		version: attempt.AuthorizationVersion,
	}, &fixtureIdentityLookup{match: fixtureMatch(true, 4)}, resources,
		&fixtureInvitation{}, receipts)
	result, err := worker.ProcessBatch(context.Background(), attempt,
		[]SourceRow{fixtureSourceRow(t, publicresource.Roles)})
	if err != nil || result.Committed != 1 || resources.updateCalls != 0 ||
		len(receipts.appended) != 1 || receipts.appended[0].Outcome != "skipped" {
		t.Fatalf("result=%+v updates=%d receipts=%+v err=%v",
			result, resources.updateCalls, receipts.appended, err)
	}
}
