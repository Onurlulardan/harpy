package catalogquery

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestResolvePlatformCatalogReturnsExactTypedSnapshots(t *testing.T) {
	offeringID := uuid.MustParse("019c92a0-0000-7000-8000-000000000001")
	discountID := uuid.MustParse("019c92a0-0000-7000-8000-000000000003")
	campaignID := uuid.MustParse("019c92a0-0000-7000-8000-000000000004")
	queries, err := New(fakeStore{platform: func(
		_ context.Context, query PlatformResolveStoreQuery,
	) (PlatformCatalogRecord, bool, error) {
		switch query.Kind {
		case PlatformOffering:
			record := offeringRecord("team", 1)
			record.ID, record.State, record.Public = offeringID, "retired", false
			return PlatformCatalogRecord{Offering: &record}, true, nil
		case PlatformDiscount:
			record := campaignRecord().Discounts[0]
			record.ID, record.State = discountID, "revoked"
			return PlatformCatalogRecord{Discount: &record}, true, nil
		case PlatformCampaign:
			record := campaignRecord()
			record.ID, record.State = campaignID, "retired"
			return PlatformCatalogRecord{Campaign: &record}, true, nil
		default:
			return PlatformCatalogRecord{}, false, nil
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	offering, err := queries.ResolvePlatformCatalog(context.Background(), PlatformOffering, offeringID)
	value, valid := offering.Offering()
	if err != nil || !valid || value.State() != "retired" || value.Public() {
		t.Fatalf("offering=%#v valid=%v err=%v", value, valid, err)
	}
	discount, err := queries.ResolvePlatformCatalog(context.Background(), PlatformDiscount, discountID)
	discountValue, valid := discount.Discount()
	if err != nil || !valid || discountValue.State() != "revoked" ||
		discountValue.Description() != "Summer discount" || discountValue.UsedCount() != 7 {
		t.Fatalf("discount=%#v valid=%v err=%v", discountValue, valid, err)
	}
	campaign, err := queries.ResolvePlatformCatalog(context.Background(), PlatformCampaign, campaignID)
	campaignValue, valid := campaign.Campaign()
	if err != nil || !valid || campaignValue.State() != "retired" ||
		campaignValue.Description() != "Loyal customers" || !campaignValue.AutoApply() ||
		len(campaignValue.Discounts()) != 1 {
		t.Fatalf("campaign=%#v valid=%v err=%v", campaignValue, valid, err)
	}
}

func TestResolvePlatformCatalogRejectsInvalidMissingAndMismatchedRecords(t *testing.T) {
	queries, _ := New(fakeStore{})
	if _, err := queries.ResolvePlatformCatalog(
		context.Background(), PlatformOffering, uuid.New(),
	); !errors.Is(err, ErrRequest) {
		t.Fatalf("non-v7 error = %v", err)
	}
	id := uuid.MustParse("019c92a0-0000-7000-8000-000000000002")
	if _, err := queries.ResolvePlatformCatalog(
		context.Background(), PlatformOffering, id,
	); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing error = %v", err)
	}
	mismatched, _ := New(fakeStore{platform: func(
		context.Context, PlatformResolveStoreQuery,
	) (PlatformCatalogRecord, bool, error) {
		record := campaignRecord()
		return PlatformCatalogRecord{Campaign: &record}, true, nil
	}})
	if _, err := mismatched.ResolvePlatformCatalog(
		context.Background(), PlatformOffering, id,
	); !errors.Is(err, ErrResult) {
		t.Fatalf("mismatched error = %v", err)
	}
}
