package query

import (
	"context"
	"errors"
	"testing"

	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestDeleteBoundPropagatesOptimisticConcurrencyConflict(t *testing.T) {
	store := &queryFixtureStore{delete: func(context.Context, DeleteMutation) (Snapshot, bool, error) {
		return Snapshot{}, false, ErrConflict
	}}
	_, err := queryFixtureApplication(queryFixtureOperation).deleteBound(context.Background(),
		platformtx.TenantScope{}, store, queryFixtureIdentity(), DeleteRequest{QueryID: queryFixtureID,
			ExpectedRootRevision: 2, IdempotencyKey: "query-delete-1"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("delete error = %v", err)
	}
}

func TestDeleteBoundRemainsAvailableWhenBuilderIsDisabled(t *testing.T) {
	storeCalled := false
	store := &queryFixtureStore{delete: func(
		context.Context, DeleteMutation,
	) (Snapshot, bool, error) {
		storeCalled = true
		result := queryFixtureSnapshot("draft", "none")
		result.Lifecycle = analyticsquery.LifecycleRetired
		result.RootRevision = 3
		return result, false, nil
	}}
	entitlements := &queryFixtureEntitlements{}
	application := queryFixtureApplication(queryFixtureOperation)
	application.entitlements = entitlements
	_, err := application.deleteBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), DeleteRequest{QueryID: queryFixtureID,
			ExpectedRootRevision: 2, IdempotencyKey: "query-delete-builder-off-1"})
	if err != nil || !storeCalled ||
		entitlements.request.Mutation != billingapi.MeteredResourceDelete {
		t.Fatalf("delete error=%v owner_called=%v mutation=%s",
			err, storeCalled, entitlements.request.Mutation)
	}
}

func TestDeleteBoundDoesNotWriteOwnerStateWhenBillingIsUnavailable(t *testing.T) {
	storeCalled := false
	store := &queryFixtureStore{delete: func(
		context.Context, DeleteMutation,
	) (Snapshot, bool, error) {
		storeCalled = true
		return Snapshot{}, false, nil
	}}
	application := queryFixtureApplication(queryFixtureOperation)
	application.entitlements = &queryFixtureEntitlements{err: errors.New("billing unavailable")}
	_, err := application.deleteBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), DeleteRequest{QueryID: queryFixtureID,
			ExpectedRootRevision: 2, IdempotencyKey: "query-delete-2"})
	if !errors.Is(err, ErrUnavailable) || storeCalled {
		t.Fatalf("delete error=%v owner_called=%v", err, storeCalled)
	}
}
