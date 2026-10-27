package query

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestUpdateBoundPropagatesOptimisticConcurrencyConflict(t *testing.T) {
	store := &queryFixtureStore{update: func(context.Context, UpdateMutation) (Snapshot, bool, error) {
		return Snapshot{}, false, ErrConflict
	}}
	document := queryFixtureDocument(`{}`)
	_, err := queryFixtureApplication(queryFixtureOperation).updateBound(context.Background(),
		platformtx.TenantScope{}, store, queryFixtureIdentity(), UpdateRequest{QueryID: queryFixtureID,
			ExpectedRootRevision: 1, ExpectedVersionRevision: 1,
			Name: "Orders", IdempotencyKey: "query-update-1"}, document, document, document)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("update error = %v", err)
	}
}

func TestUpdateBoundDoesNotWriteOwnerStateWhenBuilderIsDisabled(t *testing.T) {
	storeCalled := false
	store := &queryFixtureStore{update: func(
		context.Context, UpdateMutation,
	) (Snapshot, bool, error) {
		storeCalled = true
		return Snapshot{}, false, nil
	}}
	decision := queryFixtureMutationDecision(billingapi.MeteredResourceUpdate, false)
	decision.FeatureEnabled, decision.Allowed = false, false
	application := queryFixtureApplication(queryFixtureOperation)
	application.entitlements = &queryFixtureEntitlements{decision: decision}
	document := queryFixtureDocument(`{}`)
	_, err := application.updateBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), UpdateRequest{QueryID: queryFixtureID,
			ExpectedRootRevision: 1, ExpectedVersionRevision: 1,
			Name: "Orders", IdempotencyKey: "query-update-denied-1"},
		document, document, document)
	if !errors.Is(err, ErrEntitlement) || storeCalled {
		t.Fatalf("update error=%v owner_called=%v", err, storeCalled)
	}
}
