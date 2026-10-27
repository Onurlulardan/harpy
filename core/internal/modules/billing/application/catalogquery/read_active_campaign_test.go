package catalogquery

import (
	"context"
	"errors"
	"testing"
)

func TestReadActiveCampaignPreservesDeterministicImmutableDiscounts(t *testing.T) {
	storage := fakeStore{list: noList, get: noGet,
		campaign: func(_ context.Context, query CampaignStoreQuery) (CampaignRecord, bool, error) {
			if !query.Instant.Equal(fixtureInstant) {
				t.Fatalf("campaign query = %#v", query)
			}
			return campaignRecord(), true, nil
		}}
	queries, _ := New(storage)
	request, _ := NewActiveCampaignRequest(fixtureInstant)
	result, err := queries.ReadActiveCampaign(context.Background(), request)
	if err != nil || result == nil || result.Key() != "summer" ||
		result.Description() != "Loyal customers" || !result.AutoApply() ||
		len(result.Discounts()) != 1 || result.Discounts()[0].Code() != "SAVE15" {
		t.Fatalf("campaign = %#v, %v", result, err)
	}
	discounts := result.Discounts()
	basisPoints := discounts[0].BasisPoints()
	*basisPoints = 1
	discounts[0] = DiscountSnapshot{}
	if *result.Discounts()[0].BasisPoints() != 1500 || result.Discounts()[0].ID() == [16]byte{} {
		t.Fatal("campaign snapshot leaked mutable discount state")
	}
	if result.Discounts()[0].UsedCount() != 7 ||
		result.Discounts()[0].Name() != "Save fifteen" {
		t.Fatal("campaign lost typed discount presentation")
	}
}

func TestReadActiveCampaignAllowsEmptyAndRejectsUnorderedDiscounts(t *testing.T) {
	request, _ := NewActiveCampaignRequest(fixtureInstant)
	empty, _ := New(fakeStore{list: noList, get: noGet, campaign: noCampaign})
	if result, err := empty.ReadActiveCampaign(context.Background(), request); err != nil || result != nil {
		t.Fatalf("empty campaign = %#v, %v", result, err)
	}
	record := campaignRecord()
	second := record.Discounts[0]
	second.ID[15], second.VersionID[15], second.Priority = 31, 31, 5
	record.Discounts = append(record.Discounts, second)
	invalid, _ := New(fakeStore{list: noList, get: noGet,
		campaign: func(context.Context, CampaignStoreQuery) (CampaignRecord, bool, error) {
			return record, true, nil
		}})
	if _, err := invalid.ReadActiveCampaign(context.Background(), request); !errors.Is(err, ErrResult) {
		t.Fatalf("unordered campaign error = %v", err)
	}
}
