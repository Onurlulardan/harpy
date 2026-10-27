package customerlifecycle

import (
	"context"
	"errors"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func TestListUsesPersistedSaaSBoundaryAndBoundedQuery(t *testing.T) {
	service, access, store, _ := newFixtureService(t)
	store.page = Page{}
	page, err := service.List(context.Background(), ListQuery{Limit: 25})
	if err != nil || page.Items == nil {
		t.Fatalf("list result=%+v err=%v", page, err)
	}
	if len(access.requests) != 1 || access.requests[0].Capability != CapabilityRead ||
		access.requests[0].ExpectedMode != "saas" || access.requests[0].RequireStepUp {
		t.Fatalf("access request=%+v", access.requests)
	}
	if store.listRequest.EnvironmentID != fixtureEnvironmentID || store.listRequest.Limit != 25 {
		t.Fatalf("store request=%+v", store.listRequest)
	}
}

func TestOnPremiseModeDeniesBeforeIAMAndStorage(t *testing.T) {
	access := &fixtureAccess{now: fixtureNow}
	store := &fixtureStore{}
	service, err := New(Dependencies{
		Policy: fixturePolicy{mode: tenantapi.DeploymentModeOnPremise}, Access: access,
		Store: store, PurgeReadiness: &fixtureReadiness{}, Clock: fixtureClock{fixtureNow},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.List(context.Background(), ListQuery{})
	if !errors.Is(err, ErrModeDenied) || len(access.requests) != 0 ||
		store.listRequest.EnvironmentID != fixtureEnvironmentID && store.listRequest.EnvironmentID.String() != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("err=%v access=%d store=%+v", err, len(access.requests), store.listRequest)
	}
}

func TestControlTenantCannotBeReadAsCustomer(t *testing.T) {
	service, access, _, _ := newFixtureService(t)
	_, err := service.Get(context.Background(), GetQuery{TenantID: fixtureControlID})
	if !errors.Is(err, ErrNotFound) || len(access.requests) != 0 {
		t.Fatalf("err=%v access=%d", err, len(access.requests))
	}
}

func TestUpdateAuthorizesOnceAndPinsActorRevisionAndDigest(t *testing.T) {
	service, access, store, _ := newFixtureService(t)
	result, err := service.Update(context.Background(), MutationCommand{
		TenantID: fixtureTargetID, ExpectedRevision: 7,
		IdempotencyKey: "tenant-update-key-0001", Name: "Renamed customer",
	})
	if err != nil || result.Revision != 2 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	request := store.mutationRequest
	if len(access.requests) != 1 || request.ActorID != fixtureActorID ||
		request.Kind != MutationUpdate || request.ExpectedRevision != 7 ||
		request.RequestSHA256 == [32]byte{} || request.Now != fixtureNow {
		t.Fatalf("access=%d request=%+v", len(access.requests), request)
	}
}

func TestMutationOutcomesRemainDistinct(t *testing.T) {
	cases := []struct {
		outcome string
		want    error
	}{
		{"not_found", ErrNotFound}, {"revision_conflict", ErrRevisionConflict},
		{"idempotency_conflict", ErrIdempotencyConflict}, {"state_conflict", ErrStateConflict},
	}
	for _, test := range cases {
		t.Run(test.outcome, func(t *testing.T) {
			service, _, store, _ := newFixtureService(t)
			store.mutation.Outcome = test.outcome
			_, err := service.Suspend(context.Background(), MutationCommand{
				TenantID: fixtureTargetID, ExpectedRevision: 1,
				IdempotencyKey: "tenant-suspend-key-01",
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("err=%v want=%v", err, test.want)
			}
		})
	}
}
