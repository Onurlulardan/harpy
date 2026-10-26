package billingv1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type subscriptionPageStub struct {
	billingapi.PlatformReports
	pages    []billingapi.Page[billingapi.SubscriptionView]
	requests []billingapi.Query
}

func (stub *subscriptionPageStub) ListSubscriptions(
	_ context.Context,
	query billingapi.Query,
) (billingapi.Page[billingapi.SubscriptionView], error) {
	stub.requests = append(stub.requests, query)
	page := stub.pages[0]
	stub.pages = stub.pages[1:]
	return page, nil
}

func TestAllSubscriptionsExhaustsPagesInV1CreatedOrder(t *testing.T) {
	next := "next-page"
	stub := &subscriptionPageStub{pages: []billingapi.Page[billingapi.SubscriptionView]{
		{Items: []billingapi.SubscriptionView{billingapi.NewSubscriptionQueryView(
			json.RawMessage(`{"id":"newer","created_at":"2026-07-17T00:00:00.000000Z"}`),
		)}, Pagination: billingapi.Pagination{Next: &next}},
		{Items: []billingapi.SubscriptionView{billingapi.NewSubscriptionQueryView(
			json.RawMessage(`{"id":"older","created_at":"2026-07-16T00:00:00.000000Z"}`),
		)}},
	}}
	handler := &Handler{query: billingapi.NewQueryDecoder(), reports: stub}
	items, err := handler.allSubscriptions(context.Background(), billingapi.Access{})
	if err != nil || len(items) != 2 || items[0].ID != "newer" || items[1].ID != "older" {
		t.Fatalf("subscriptions=%+v error=%v", items, err)
	}
	if len(stub.requests) != 2 {
		t.Fatalf("subscription request count=%d", len(stub.requests))
	}
	for index, query := range stub.requests {
		sorts := query.Request.Sort()
		if len(sorts) != 1 || sorts[0].Field() != "created_at" ||
			sorts[0].Direction().String() != "desc" {
			t.Fatalf("request %d sorts=%v", index, sorts)
		}
	}
	after, found := stub.requests[1].Request.Page().After()
	if !found || after != next {
		t.Fatalf("second page cursor=%q found=%t", after, found)
	}
}

func TestSubscriptionLookupUsesBoundedExactPredicate(t *testing.T) {
	handler := &Handler{query: billingapi.NewQueryDecoder()}
	query, err := handler.subscriptionLookupQuery(
		billingapi.Access{}, "legacy_id", "11111111-1111-4111-8111-111111111111",
	)
	where, found := query.Request.Where()
	if !found {
		t.Fatal("lookup filter is missing")
	}
	children := where.Children()
	if len(children) != 1 {
		t.Fatalf("lookup filter children=%d", len(children))
	}
	predicate, predicateFound := children[0].Predicate()
	_, hasCursor := query.Request.Page().After()
	if err != nil || query.Request.Page().Size() != 2 || hasCursor ||
		!predicateFound || predicate.Field() != "legacy_id" {
		t.Fatalf("lookup page=%d filter=%v error=%v",
			query.Request.Page().Size(), predicate, err)
	}
}

func TestSubscriptionLookupFailsClosedOnDuplicateInOneQuery(t *testing.T) {
	id := "11111111-1111-4111-8111-111111111111"
	stub := &subscriptionPageStub{pages: []billingapi.Page[billingapi.SubscriptionView]{
		{Items: []billingapi.SubscriptionView{{}, {}}},
	}}
	handler := &Handler{query: billingapi.NewQueryDecoder(), reports: stub}
	request := httptest.NewRequest("GET", "/api/v1/billing/subscription/"+id, nil)

	_, err := handler.findSubscription(request, billingapi.Access{}, id)
	if !errors.Is(err, ErrSnapshotRequired) {
		t.Fatalf("duplicate lookup error=%v", err)
	}
	if len(stub.requests) != 1 {
		t.Fatalf("duplicate lookup request count=%d", len(stub.requests))
	}
	if stub.requests[0].Request.Page().Size() != 2 {
		t.Fatalf("duplicate lookup page size=%d", stub.requests[0].Request.Page().Size())
	}
	if _, found := stub.requests[0].Request.Page().After(); found {
		t.Fatal("first lookup request unexpectedly carries a cursor")
	}
}
