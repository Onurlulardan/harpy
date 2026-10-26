package catalogquery

import (
	"context"
	"time"

	"github.com/google/uuid"
)

var fixtureInstant = time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)

type fakeStore struct {
	list     func(context.Context, ListStoreQuery) ([]OfferingRecord, error)
	get      func(context.Context, GetStoreQuery) (OfferingRecord, bool, error)
	campaign func(context.Context, CampaignStoreQuery) (CampaignRecord, bool, error)
	platform func(
		context.Context, PlatformResolveStoreQuery,
	) (PlatformCatalogRecord, bool, error)
	legacy func(
		context.Context, LegacyIdentityStoreQuery,
	) (LegacyIdentityRecord, bool, error)
}

func (fake fakeStore) ResolvePlatformCatalog(
	ctx context.Context, query PlatformResolveStoreQuery,
) (PlatformCatalogRecord, bool, error) {
	if fake.platform == nil {
		return PlatformCatalogRecord{}, false, nil
	}
	return fake.platform(ctx, query)
}

func (fake fakeStore) MaterializeLegacyIdentity(
	ctx context.Context, query LegacyIdentityStoreQuery,
) (LegacyIdentityRecord, bool, error) {
	if fake.legacy != nil {
		return fake.legacy(ctx, query)
	}
	return LegacyIdentityRecord{
		LegacyID: query.LegacyID, CanonicalID: query.LegacyID, Kind: query.Kind,
	}, true, nil
}

func (fake fakeStore) ListOfferings(
	ctx context.Context,
	query ListStoreQuery,
) ([]OfferingRecord, error) {
	return fake.list(ctx, query)
}

func (fake fakeStore) GetOffering(
	ctx context.Context,
	query GetStoreQuery,
) (OfferingRecord, bool, error) {
	return fake.get(ctx, query)
}

func (fake fakeStore) ReadActiveCampaign(
	ctx context.Context,
	query CampaignStoreQuery,
) (CampaignRecord, bool, error) {
	return fake.campaign(ctx, query)
}

func offeringRecord(key string, seed byte) OfferingRecord {
	value := true
	return OfferingRecord{
		ID: uuid.UUID{15: seed}, Key: key, Revision: 2,
		VersionID: uuid.UUID{14: 1, 15: seed}, Version: 2,
		State: "active", Public: true, Available: true, Name: "Public " + key,
		Description: "Description " + key, SortOrder: int32(seed),
		HighlightLabel: "Popular", Color: "blue", BillingInterval: "monthly",
		Base: seed == 1, TrialDays: 14, GraceDays: 3,
		ExpectedPriceCount: 1, ExpectedGrantCount: 1,
		EffectiveStart: fixtureInstant.Add(-time.Hour),
		CreatedAt:      fixtureInstant.Add(-24 * time.Hour), UpdatedAt: fixtureInstant.Add(-time.Minute),
		Prices: []PriceRecord{{ID: uuid.UUID{13: 1, 15: seed}, Cadence: "monthly",
			MinorUnits: 9900, Currency: "USD", CurrencyRegistryVersion: 1}},
		Entitlements: []EntitlementRecord{{Key: "feature.enabled", Kind: "boolean",
			Boolean: &value}},
	}
}

func campaignRecord() CampaignRecord {
	basisPoints := int64(1500)
	record := CampaignRecord{
		ID: uuid.UUID{15: 20}, Key: "summer", Revision: 4,
		VersionID: uuid.UUID{14: 1, 15: 20}, Version: 3,
		Name: "Summer", Description: "Loyal customers", BannerText: "Save now",
		AutoApply: true, Available: true, Priority: 100,
		StartsAt:  fixtureInstant.Add(-time.Hour),
		CreatedAt: fixtureInstant.Add(-24 * time.Hour),
		UpdatedAt: fixtureInstant.Add(-time.Minute),
		Discounts: []DiscountRecord{{
			ID: uuid.UUID{15: 30}, Code: "SAVE15", Revision: 2,
			VersionID: uuid.UUID{14: 1, 15: 30}, Version: 2,
			Name: "Save fifteen", Description: "Summer discount", Available: true,
			Kind: "basis_point", BasisPoints: &basisPoints, UsedCount: 7,
			Duration: "once", Priority: 10, AllOfferings: true,
			ValidStart: fixtureInstant.Add(-time.Hour),
			CreatedAt:  fixtureInstant.Add(-24 * time.Hour),
			UpdatedAt:  fixtureInstant.Add(-time.Minute),
		}},
	}
	record.LegacyID = record.ID
	record.Discounts[0].LegacyID = record.Discounts[0].ID
	return record
}

func noList(context.Context, ListStoreQuery) ([]OfferingRecord, error) { return nil, nil }
func noGet(context.Context, GetStoreQuery) (OfferingRecord, bool, error) {
	return OfferingRecord{}, false, nil
}
func noCampaign(context.Context, CampaignStoreQuery) (CampaignRecord, bool, error) {
	return CampaignRecord{}, false, nil
}
