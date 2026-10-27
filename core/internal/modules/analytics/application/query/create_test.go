package query

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateBoundPreservesReplayAndCanonicalMutation(t *testing.T) {
	var captured CreateMutation
	port := &queryFixtureEntitlements{decision: queryCreateDecision(true)}
	store := &queryFixtureStore{create: func(_ context.Context, mutation CreateMutation) (Snapshot, bool, error) {
		captured = mutation
		return queryFixtureSnapshot("draft", "none"), true, nil
	}}
	document := queryFixtureDocument(` { } `)
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureID,
		queryFixtureVersion)
	application.entitlements = port
	result, err := application.createBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(),
		CreateRequest{Name: "Orders", IdempotencyKey: "query-create-1"},
		document, document, document)
	if err != nil || result.QueryID != queryFixtureID || captured.RequestSHA256 == [32]byte{} ||
		captured.OperationID != queryFixtureOperation || string(captured.Query.CanonicalJSON()) != `{}` {
		t.Fatalf("create replay = %+v, mutation = %+v, err = %v", result, captured, err)
	}
}

func TestCreateBoundRequiresExactBillingAdmissionForNewQuery(t *testing.T) {
	fixture := &queryFixtureEntitlements{}
	store := &queryFixtureStore{create: func(
		_ context.Context, _ CreateMutation,
	) (Snapshot, bool, error) {
		if !fixture.called {
			t.Fatal("query owner store called before Billing decision")
		}
		return queryFixtureSnapshot("draft", "none"), false, nil
	}}
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureID,
		queryFixtureVersion)
	application.entitlements = fixture
	document := queryFixtureDocument(`{}`)
	_, err := application.createBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), CreateRequest{Name: "Orders",
			IdempotencyKey: "query-create-2"}, document, document, document)
	if err != nil || !fixture.called {
		t.Fatalf("query admission called=%v err=%v", fixture.called, err)
	}
}

func TestCreateBoundFailsClosedForBillingDenialMismatchAndUnavailable(t *testing.T) {
	document := queryFixtureDocument(`{}`)
	tests := []struct {
		name string
		port *queryFixtureEntitlements
		want error
	}{
		{name: "denied", port: &queryFixtureEntitlements{decision: billingapi.MeteredResourceMutationDecision{
			Resource: billingapi.MeteredResourceAPIEndpoint,
			Mutation: billingapi.MeteredResourceCreate,
			TenantID: queryFixtureTenant, OperationID: queryFixtureOperation,
			OwnerFactID: queryFixtureID, FeatureEnabled: true,
			Current: 1, Limit: 1,
		}}, want: ErrEntitlement},
		{name: "mismatch", port: &queryFixtureEntitlements{decision: billingapi.MeteredResourceMutationDecision{
			Resource: billingapi.MeteredResourceAPIEndpoint,
			Mutation: billingapi.MeteredResourceCreate,
			TenantID: queryFixtureTenant, OperationID: queryFixtureOperation,
			OwnerFactID: queryFixtureNext, FeatureEnabled: true,
			Allowed: true, GaugeDelta: 1, Unlimited: true,
		}}, want: ErrUnavailable},
		{name: "unavailable", port: &queryFixtureEntitlements{
			err: errors.New("billing unavailable")}, want: ErrUnavailable},
		{name: "billing conflict", port: &queryFixtureEntitlements{
			err: billingapi.ErrMeteredResourceIdempotencyReused}, want: ErrConflict},
		{name: "billing invalid", port: &queryFixtureEntitlements{
			err: billingapi.ErrMeteredResourceRequest}, want: ErrInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ownerCalled := false
			store := &queryFixtureStore{create: func(
				_ context.Context, _ CreateMutation,
			) (Snapshot, bool, error) {
				ownerCalled = true
				return queryFixtureSnapshot("draft", "none"), false, nil
			}}
			application := queryFixtureApplication(queryFixtureOperation,
				queryFixtureID, queryFixtureVersion)
			application.entitlements = test.port
			_, err := application.createBound(context.Background(),
				platformtx.TenantScope{}, store, queryFixtureIdentity(),
				CreateRequest{Name: "Orders", IdempotencyKey: "query-create-3"},
				document, document, document)
			if !errors.Is(err, test.want) {
				t.Fatalf("create error = %v, want %v", err, test.want)
			}
			if ownerCalled {
				t.Fatal("query owner store called after Billing denied or failed closed")
			}
		})
	}
}

func TestCreateBoundPropagatesOwnerFailureAfterBillingAllow(t *testing.T) {
	port := &queryFixtureEntitlements{}
	store := &queryFixtureStore{create: func(
		_ context.Context, _ CreateMutation,
	) (Snapshot, bool, error) {
		if !port.called {
			t.Fatal("query owner store called before Billing decision")
		}
		return Snapshot{}, false, ErrConflict
	}}
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureID,
		queryFixtureVersion)
	application.entitlements = port
	document := queryFixtureDocument(`{}`)
	_, err := application.createBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), CreateRequest{Name: "Orders",
			IdempotencyKey: "query-create-4"}, document, document, document)
	if !errors.Is(err, ErrConflict) || !port.called {
		t.Fatalf("create owner failure called=%v err=%v", port.called, err)
	}
}

func queryCreateDecision(replay bool) billingapi.MeteredResourceMutationDecision {
	return queryFixtureMutationDecision(billingapi.MeteredResourceCreate, replay)
}
