package billing

import (
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

func currentSubscription(
	t *testing.T,
	tenantID uuid.UUID,
) billingapi.TenantSubscriptionSnapshot {
	t.Helper()
	now := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	identifier, offeringID, priceID := testV7(t), testV7(t), testV7(t)
	rootID := testV7(t)
	offering := billingapi.OfferingView{
		ID: rootID.String(), Key: "historical-team", Revision: 2,
		VersionID: offeringID.String(), Version: 1, State: "retired",
		Name: "Historical Team", BillingInterval: "monthly", Public: false,
		EffectiveStart: formatInstant(now.Add(-24 * time.Hour)),
		Prices: []billingapi.PriceView{{ID: priceID.String(), Cadence: "monthly",
			Amount:  billingapi.MoneyView{MinorUnits: 9900, Currency: "USD", RegistryVersion: 1},
			Version: 1}},
		Entitlements: []billingapi.EntitlementValueView{},
		CreatedAt:    now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour),
	}
	view := billingapi.SubscriptionView{
		ID: identifier.String(), TenantID: tenantID.String(),
		OfferingVersionID: offeringID.String(), PriceVersionID: priceID.String(),
		Provider: "manual", State: "active",
		PeriodStart: formatInstant(now.Add(-time.Hour)),
		PeriodEnd:   formatInstant(now.AddDate(0, 1, 0)), Version: 3,
		EffectiveMinor: 9900, EffectiveCurrency: "USD", EffectiveRegistry: 1,
		CreatedAt: formatInstant(now.Add(-48 * time.Hour)), UpdatedAt: formatInstant(now),
		Offering: offering,
	}
	return billingapi.TenantSubscriptionSnapshot{
		ID: identifier, TenantID: tenantID,
		OfferingVersionID: offeringID, PriceVersionID: priceID, View: view,
	}
}

func revisedSubscription(
	value billingapi.TenantSubscriptionSnapshot,
	cancel bool,
) billingapi.TenantSubscriptionSnapshot {
	result := value
	result.View.Version++
	result.View.CancelAtPeriodEnd = cancel
	return result
}

func usageFixture(tenantID uuid.UUID) billingapi.TenantUsageSnapshot {
	limit := int64(100)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	periodStart, periodEnd := formatInstant(start), formatInstant(start.AddDate(0, 1, 0))
	return billingapi.TenantUsageSnapshot{
		TenantID: tenantID,
		View: billingapi.UsageView{Meters: []billingapi.UsageMeterView{{
			EntitlementKey: "forms.submissions", Kind: "integer",
			Current: 9, Limit: &limit, ResetPolicy: "calendar_month_utc",
			PeriodStart: &periodStart, PeriodEnd: &periodEnd,
		}}},
	}
}

func eligibilitySnapshot(
	t *testing.T,
	tenantID uuid.UUID,
	current billingapi.TenantSubscriptionSnapshot,
) billingapi.TenantDiscountSnapshot {
	t.Helper()
	discountID, versionID := testV7(t), testV7(t)
	return billingapi.TenantDiscountSnapshot{
		TenantID: tenantID, DiscountID: discountID, DiscountVersionID: versionID,
		SubscriptionID: current.ID, OfferingVersionID: current.OfferingVersionID,
		PriceVersionID: current.PriceVersionID, Currency: "USD",
		View: billingapi.DiscountView{
			ID: discountID.String(), Code: "LAUNCH", Revision: 1,
			VersionID: versionID.String(), Version: 1, State: "active",
			Kind: "basis_points", BasisPoints: int64Pointer(1500), Duration: "once",
			Priority: 2, AllOfferings: true, OfferingIDs: []string{},
			ValidStart: formatInstant(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)),
		},
	}
}

func int64Pointer(value int64) *int64 { return &value }
