//go:build integration

package integrationtest

import (
	"strings"
	"time"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
	"github.com/google/uuid"
)

func billingV1AdoptionFixture(at time.Time) (
	billingv1adoptionapp.AdoptionSnapshot,
	billingv1adoptionapp.AdoptionMapping,
) {
	environmentID := uuid.MustParse(billingEnvironmentID)
	tenantID := uuid.MustParse(billingTenantID)
	planID := uuid.MustParse("72000000-0000-4000-8000-000000000072")
	inactivePlanID := uuid.MustParse("72100000-0000-4000-8000-000000000072")
	deletedPlanID := uuid.MustParse("72200000-0000-4000-8000-000000000072")
	campaignID := uuid.MustParse("73000000-0000-4000-8000-000000000073")
	couponID := uuid.MustParse("74000000-0000-4000-8000-000000000074")
	subscriptionID := uuid.MustParse("75000000-0000-4000-8000-000000000075")
	usageID := uuid.MustParse("77000000-0000-4000-8000-000000000077")
	eventID := uuid.MustParse("7b000000-0000-4000-8000-00000000007b")
	created := at.AddDate(-1, 0, 0)
	updated := created.AddDate(0, 0, 1)
	deleted := created.AddDate(0, 1, 0)
	snapshot := billingv1adoptionapp.AdoptionSnapshot{Global: billingv1adoptionapp.GlobalSnapshot{
		EnvironmentID: environmentID, SnapshotAt: at,
		Plans: []billingv1adoptionapp.SourcePlan{{
			ID: planID, Name: "Base", Slug: "base-plan", PriceCents: 1000,
			AnnualPriceCents: 10000, BillingInterval: "monthly", Currency: "USD",
			LimitsJSON:   []byte(`{"submissions_per_month":100,"max_users":10,"max_forms":20,"max_storage_mb":50}`),
			FeaturesJSON: []byte(`{"ai_chat":true,"workflows":true,"api_access":true}`),
			IsActive:     true, IsPublic: true, CreatedAt: &created, UpdatedAt: &created,
		}, {
			ID: inactivePlanID, Name: "Hidden", Slug: "hidden-plan", PriceCents: 2000,
			BillingInterval: "monthly", Currency: "USD",
			LimitsJSON:   []byte(`{"submissions_per_month":10,"max_users":2,"max_forms":3,"max_storage_mb":5}`),
			FeaturesJSON: []byte(`{"ai_chat":false,"workflows":false,"api_access":false}`),
			IsActive:     false, IsPublic: true, CreatedAt: &created, UpdatedAt: &created,
		}, {
			ID: deletedPlanID, Name: "Deleted", Slug: "deleted-plan", PriceCents: 3000,
			BillingInterval: "monthly", Currency: "USD",
			LimitsJSON:   []byte(`{"submissions_per_month":10,"max_users":2,"max_forms":3,"max_storage_mb":5}`),
			FeaturesJSON: []byte(`{"ai_chat":false,"workflows":false,"api_access":false}`),
			IsActive:     false, IsPublic: true, CreatedAt: &created, UpdatedAt: &deleted,
			DeletedAt: &deleted,
		}},
		Campaigns: []billingv1adoptionapp.SourceCampaign{{
			ID:        campaignID,
			CreatedBy: uuid.MustParse("73100000-0000-4000-8000-000000000073"),
			Name:      "Launch", Slug: "launch-sale", Description: "Launch campaign",
			StartsAt: created, AutoApply: true, BannerText: "Save now",
			IsActive: false, CreatedAt: &created, UpdatedAt: &updated,
		}},
		Coupons: []billingv1adoptionapp.SourceCoupon{{
			ID:        couponID,
			CreatedBy: uuid.MustParse("74100000-0000-4000-8000-000000000074"),
			Code:      "SAVE10", Name: "Save 10", Description: "Save ten percent",
			DiscountType: "percentage", DiscountValue: 10, UsedCount: 1,
			MaxUsesPerTenant: 1, ValidFrom: created,
			ApplicablePlanIDs: []uuid.UUID{planID}, Duration: "once",
			IsActive: false, CampaignID: &campaignID, CreatedAt: &created, UpdatedAt: &updated,
		}},
	}}
	snapshot.Tenants = []billingv1adoptionapp.TenantSnapshot{{
		TenantID: tenantID, SnapshotAt: at,
		Subscriptions: []billingv1adoptionapp.SourceSubscription{{
			ID: subscriptionID, TenantID: tenantID, PlanID: planID, State: "active",
			PeriodStart: at.AddDate(0, 0, -2), PeriodEnd: at.AddDate(0, 1, -2),
			EffectivePriceCents: 900, CreatedAt: &created, UpdatedAt: &at,
		}},
		Redemptions: []billingv1adoptionapp.SourceRedemption{{
			ID:             uuid.MustParse("76000000-0000-4000-8000-000000000076"),
			SubscriptionID: subscriptionID, CouponID: couponID, TenantID: tenantID,
			DiscountAmountCents: 100, AppliedAt: at.AddDate(0, 0, -1), CreatedAt: &at,
		}},
		Usage: []billingv1adoptionapp.SourceUsage{{
			ID: usageID, TenantID: tenantID, Month: at.Format("2006-01"),
			Count: 5, CreatedAt: &at, UpdatedAt: &at,
		}},
		Gauges: billingV1GaugeFixture(at),
		Events: []billingv1adoptionapp.SourceBillingEvent{{
			ID: eventID, TenantID: tenantID, SubscriptionID: &subscriptionID,
			EventType: "payment_succeeded", AmountCents: 900, Currency: "USD",
			Description: "paid", Metadata: []byte(`{"receipt":"R-1"}`),
			CreatedAt: at.Add(-time.Hour),
		}},
	}}
	mapping := billingv1adoptionapp.AdoptionMapping{
		ContractVersion: billingv1adoptionapp.MappingContractVersion,
		SourceTimezone:  "UTC",
		Global:          billingv1adoptionapp.GlobalMapping{BasePlanID: planID},
		Tenants: []billingv1adoptionapp.TenantMapping{{
			TenantID:                 tenantID,
			UsageFlushEvidenceSHA256: map[uuid.UUID]string{usageID: strings.Repeat("11", 32)},
		}},
	}
	return snapshot, mapping
}

func billingV1GaugeFixture(at time.Time) []billingv1adoptionapp.GaugeOpening {
	keys := []string{"documents.storage_bytes", "forms.active_forms", "iam.active_users"}
	result := make([]billingv1adoptionapp.GaugeOpening, 0, len(keys))
	for index, key := range keys {
		result = append(result, billingv1adoptionapp.GaugeOpening{
			EntitlementKey: key, Value: int64(index + 1), OwnerFactID: billingV7Fixture(index),
			EvidenceSHA256: strings.Repeat("22", 32), MeasuredAt: at,
		})
	}
	return result
}

func billingV7Fixture(index int) uuid.UUID {
	values := []string{
		"78000000-0000-4000-8000-000000000078",
		"79000000-0000-4000-8000-000000000079",
		"7a000000-0000-4000-8000-00000000007a",
	}
	return uuid.MustParse(values[index])
}
