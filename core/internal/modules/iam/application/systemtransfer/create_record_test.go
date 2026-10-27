package systemtransfer

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestCreateRecordUsesCanonicalPortForRolesAndOrganizations(t *testing.T) {
	for _, resource := range []publicresource.Resource{
		publicresource.Roles, publicresource.Organizations,
	} {
		t.Run(string(resource), func(t *testing.T) {
			attempt := fixtureAttempt(t, resource)
			row := fixtureSourceRow(t, resource)
			mutation, err := fixtureMutationResult(resource,
				publicresource.ID(fixtureRecordID), 1, false)
			if err != nil {
				t.Fatal(err)
			}
			resources := &fixtureResourcePort{result: mutation}
			worker, err := NewWorker(&fixtureScopes{}, &fixtureAuthorizer{
				tenantID: fixtureTenantID, actorID: fixtureActorID,
				version: attempt.AuthorizationVersion,
			}, &fixtureIdentityLookup{}, resources, &fixtureInvitation{},
				&fixtureReceiptStore{})
			if err != nil {
				t.Fatal(err)
			}
			result, err := worker.ProcessBatch(
				context.Background(), attempt, []SourceRow{row},
			)
			if err != nil || result.Committed != 1 || resources.prepareCalls != 1 ||
				resources.createCalls != 1 || resources.preparedResource != resource {
				t.Fatalf("result=%+v port=%+v err=%v", result, resources, err)
			}
			wantKey, _ := RowIdempotencyKey(attempt.ID, row.Index, row.SHA256)
			if resources.createKey != wantKey {
				t.Fatalf("key=%q want=%q", resources.createKey, wantKey)
			}
		})
	}
}
