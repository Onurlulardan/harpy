package billingv1adoptionapp

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCompileBuildsEligibleDeterministicBundle(t *testing.T) {
	snapshot, mapping := validAdoptionFixture()
	bundle, err := Compile(snapshot, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if !bundle.Global.Eligible || len(bundle.Tenants) != 1 || !bundle.Tenants[0].Eligible {
		t.Fatalf("unexpected eligibility: global=%v tenant=%v issues=%v/%v",
			bundle.Global.Eligible, bundle.Tenants[0].Eligible,
			bundle.Global.Issues, bundle.Tenants[0].Issues)
	}
	if len(bundle.Global.Offerings) != 1 || len(bundle.Global.Discounts) != 1 ||
		len(bundle.Global.Campaigns) != 1 || len(bundle.Tenants[0].Subscriptions) != 1 ||
		len(bundle.Tenants[0].Redemptions) != 1 || len(bundle.Tenants[0].LegacyEvents) != 1 ||
		len(bundle.Tenants[0].GaugeOpenings) != 2 {
		t.Fatal("compiled Billing V1 bundle lost canonical records")
	}
	first, firstSHA, err := EncodeScopePlan(bundle.Tenants[0])
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := Compile(snapshot, mapping)
	if err != nil {
		t.Fatal(err)
	}
	second, secondSHA, err := EncodeScopePlan(repeated.Tenants[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) || firstSHA != secondSHA {
		t.Fatal("repeated Billing V1 compilation changed the plan")
	}
}

func validAdoptionFixture() (AdoptionSnapshot, AdoptionMapping) {
	at := time.Date(2026, 7, 17, 10, 0, 0, 123000, time.UTC)
	environmentID := uuid.MustParse("70000000-0000-4000-8000-000000000007")
	tenantID := uuid.MustParse("71000000-0000-4000-8000-000000000071")
	planID := uuid.MustParse("72000000-0000-4000-8000-000000000072")
	campaignID := uuid.MustParse("73000000-0000-4000-8000-000000000073")
	couponID := uuid.MustParse("74000000-0000-4000-8000-000000000074")
	subscriptionID := uuid.MustParse("75000000-0000-4000-8000-000000000075")
	redemptionID := uuid.MustParse("76000000-0000-4000-8000-000000000076")
	usageID := uuid.MustParse("77000000-0000-4000-8000-000000000077")
	eventID := uuid.MustParse("7b000000-0000-4000-8000-00000000007b")
	created := at.AddDate(-1, 0, 0)
	snapshot := AdoptionSnapshot{Global: GlobalSnapshot{
		EnvironmentID: environmentID, SnapshotAt: at,
		Plans: []SourcePlan{{ID: planID, Name: "Base", Slug: "base-plan",
			PriceCents: 1000, AnnualPriceCents: 10000, BillingInterval: "monthly",
			Currency: "USD", LimitsJSON: []byte(`{"submissions_per_month":100,"max_users":10,"max_forms":20,"max_storage_mb":50}`),
			FeaturesJSON: []byte(`{"workflows":true,"api_access":true}`),
			IsActive:     true, IsPublic: true, TrialDays: 0, CreatedAt: &created, UpdatedAt: &created}},
		Campaigns: []SourceCampaign{{ID: campaignID, Name: "Launch", Slug: "launch-sale",
			StartsAt: created, IsActive: true, CreatedAt: &created, UpdatedAt: &created}},
		Coupons: []SourceCoupon{{ID: couponID, Code: "SAVE10", Name: "Save 10",
			DiscountType: "percentage", DiscountValue: 10, UsedCount: 1,
			MaxUsesPerTenant: 1, ValidFrom: created, ApplicablePlanIDs: []uuid.UUID{planID},
			Duration: "once", IsActive: true, CampaignID: &campaignID,
			CreatedAt: &created, UpdatedAt: &created}},
	}}
	snapshot.Tenants = []TenantSnapshot{{TenantID: tenantID, SnapshotAt: at,
		Subscriptions: []SourceSubscription{{ID: subscriptionID, TenantID: tenantID,
			PlanID: planID, State: "active", PeriodStart: at.AddDate(0, 0, -2),
			PeriodEnd: at.AddDate(0, 1, -2), EffectivePriceCents: 900,
			CreatedAt: &created, UpdatedAt: &at}},
		Redemptions: []SourceRedemption{{ID: redemptionID, SubscriptionID: subscriptionID,
			CouponID: couponID, TenantID: tenantID, DiscountAmountCents: 100,
			AppliedAt: at.AddDate(0, 0, -1), CreatedAt: &at}},
		Usage: []SourceUsage{{ID: usageID, TenantID: tenantID, Month: "2026-07",
			Count: 5, CreatedAt: &at, UpdatedAt: &at}},
		Gauges: validGauges(at),
		Events: []SourceBillingEvent{{
			ID: eventID, TenantID: tenantID, SubscriptionID: &subscriptionID,
			EventType: "payment_succeeded", AmountCents: 900, Currency: "USD",
			Description: "legacy payment", Metadata: []byte(`{"receipt":"R-1"}`),
			CreatedAt: at,
		}},
	}}
	mapping := AdoptionMapping{ContractVersion: MappingContractVersion,
		SourceTimezone: "UTC", Global: GlobalMapping{BasePlanID: planID},
		Tenants: []TenantMapping{{TenantID: tenantID,
			UsageFlushEvidenceSHA256: map[uuid.UUID]string{usageID: strings.Repeat("11", 32)}}}}
	return snapshot, mapping
}

func validGauges(at time.Time) []GaugeOpening {
	result := make([]GaugeOpening, 0, 3)
	for position, key := range requiredGaugeKeys {
		result = append(result, GaugeOpening{EntitlementKey: key, Value: int64(position + 1),
			OwnerFactID: uuid.MustParse([]string{
				"78000000-0000-4000-8000-000000000078",
				"79000000-0000-4000-8000-000000000079",
				"7a000000-0000-4000-8000-00000000007a",
			}[position]), EvidenceSHA256: strings.Repeat("22", 32), MeasuredAt: at})
	}
	return result
}
