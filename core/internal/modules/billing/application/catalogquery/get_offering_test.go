package catalogquery

import (
	"context"
	"errors"
	"testing"
)

func TestGetOfferingRequiresExactPublicKeySnapshot(t *testing.T) {
	storage := fakeStore{list: noList, campaign: noCampaign,
		get: func(_ context.Context, query GetStoreQuery) (OfferingRecord, bool, error) {
			if query.Key != "starter" || !query.Instant.Equal(fixtureInstant) {
				t.Fatalf("store query = %#v", query)
			}
			return offeringRecord("starter", 1), true, nil
		}}
	queries, _ := New(storage)
	request, _ := BuildOfferingRequest("starter", fixtureInstant)
	result, err := queries.GetOffering(context.Background(), request)
	if err != nil || result.Key() != "starter" || !result.Public() ||
		result.Name() != "Public starter" {
		t.Fatalf("offering = %#v, %v", result, err)
	}
}

func TestGetOfferingDistinguishesMissingFromInvalidProjection(t *testing.T) {
	request, _ := BuildOfferingRequest("starter", fixtureInstant)
	missing, _ := New(fakeStore{list: noList, campaign: noCampaign, get: noGet})
	if _, err := missing.GetOffering(context.Background(), request); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing error = %v", err)
	}
	invalid, _ := New(fakeStore{list: noList, campaign: noCampaign,
		get: func(context.Context, GetStoreQuery) (OfferingRecord, bool, error) {
			return offeringRecord("different", 1), true, nil
		}})
	if _, err := invalid.GetOffering(context.Background(), request); !errors.Is(err, ErrResult) {
		t.Fatalf("mismatched projection error = %v", err)
	}
}
