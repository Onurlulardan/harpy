package catalogquery

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestListOfferingsIsKeyOrderedPagedAndImmutable(t *testing.T) {
	storage := fakeStore{get: noGet, campaign: noCampaign}
	storage.list = func(_ context.Context, query ListStoreQuery) ([]OfferingRecord, error) {
		if query.AfterKey != "" || query.Limit != 3 || !query.Instant.Equal(fixtureInstant) {
			t.Fatalf("store query = %#v", query)
		}
		return []OfferingRecord{
			offeringRecord("alpha", 1), offeringRecord("bravo", 2),
			offeringRecord("charlie", 3),
		}, nil
	}
	queries, err := New(storage)
	if err != nil {
		t.Fatal(err)
	}
	request, _ := BuildOfferingListRequest(2, "", fixtureInstant)
	page, err := queries.ListOfferings(context.Background(), request)
	if err != nil || len(page.Items()) != 2 || page.Items()[0].Key() != "alpha" ||
		page.Items()[0].Description() != "Description alpha" ||
		page.Items()[0].BillingInterval() != "monthly" ||
		!page.Items()[0].CreatedAt().Equal(fixtureInstant.Add(-24*time.Hour)) ||
		page.Items()[1].Key() != "bravo" || page.NextCursor() == "" {
		t.Fatalf("page = %#v, %v", page, err)
	}
	next, err := BuildOfferingListRequest(2, page.NextCursor(), fixtureInstant)
	if err != nil || next.after != "bravo" {
		t.Fatalf("next cursor = %#v, %v", next, err)
	}
	items := page.Items()
	items[0] = OfferingSnapshot{}
	prices := page.Items()[0].Prices()
	prices[0] = PriceSnapshot{}
	boolean := page.Items()[0].Entitlements()[0].Boolean()
	*boolean = false
	if page.Items()[0].ID() == [16]byte{} || page.Items()[0].Prices()[0].ID() == [16]byte{} ||
		!*page.Items()[0].Entitlements()[0].Boolean() {
		t.Fatal("public offering snapshot leaked mutable backing state")
	}
}

func TestListOfferingsRejectsNonDeterministicStoreOrder(t *testing.T) {
	storage := fakeStore{get: noGet, campaign: noCampaign,
		list: func(context.Context, ListStoreQuery) ([]OfferingRecord, error) {
			return []OfferingRecord{
				offeringRecord("bravo", 2), offeringRecord("alpha", 1),
			}, nil
		}}
	queries, _ := New(storage)
	request, _ := BuildOfferingListRequest(2, "", fixtureInstant)
	if _, err := queries.ListOfferings(context.Background(), request); !errors.Is(err, ErrResult) {
		t.Fatalf("unordered store result error = %v", err)
	}
}
