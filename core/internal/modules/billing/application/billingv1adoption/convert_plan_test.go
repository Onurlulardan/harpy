package billingv1adoptionapp

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

func TestConvertPlanMapsExactTypedCatalog(t *testing.T) {
	request := ConversionRequest{Snapshot: PlanSnapshot{
		Name: "Pro", Slug: "pro", PriceCents: 9900, AnnualPriceCents: 99000,
		BillingInterval: "monthly", Currency: "USD", IsPublic: true, TrialDays: 14,
		LimitsJSON:   []byte(`{"submissions_per_month":500,"max_users":3,"max_forms":10,"max_storage_mb":100}`),
		FeaturesJSON: []byte(`{"ai_chat":true,"workflows":false,"api_access":true}`),
	}, IsBase: false, Available: true, SnapshotAt: snapshotInstant()}
	result, err := ConvertPlan(request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Key != "pro" || len(result.MissingFields) != 0 ||
		len(result.Version.Prices()) != 2 || len(result.Version.Grants()) != 7 ||
		result.Version.TrialDays() != 14 || !result.Version.Public() ||
		!result.Version.Available() {
		t.Fatalf("result = %#v", result)
	}
	values := grantValues(result.Version)
	if value, ok := values["forms.submissions"].Integer(); !ok || value != 500 {
		t.Fatalf("submissions = %d, %t", value, ok)
	}
	if value, ok := values["documents.storage_bytes"].Integer(); !ok || value != 100*1_048_576 {
		t.Fatalf("storage = %d, %t", value, ok)
	}
	if value, ok := values["ai.chat"].Boolean(); !ok || !value {
		t.Fatalf("ai.chat = %t, %t", value, ok)
	}
	if value, ok := values["workflow.enabled"].Boolean(); !ok || value {
		t.Fatalf("workflow.enabled = %t, %t", value, ok)
	}
}

func TestConvertPlanPreservesMissingAsZeroWithEvidence(t *testing.T) {
	result, err := ConvertPlan(ConversionRequest{Snapshot: PlanSnapshot{
		Name: "Base", Slug: "base", BillingInterval: "monthly", Currency: "USD",
		LimitsJSON: []byte(`{}`), FeaturesJSON: []byte(`{}`),
	}, IsBase: true, Available: true, SnapshotAt: snapshotInstant()})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.MissingFields) != 7 || !result.Version.Base() {
		t.Fatalf("missing/base = %d/%t", len(result.MissingFields), result.Version.Base())
	}
	values := grantValues(result.Version)
	for _, key := range []entitlement.Key{
		"forms.submissions", "iam.active_users", "forms.active_forms",
		"documents.storage_bytes",
	} {
		if value, ok := values[key].Integer(); !ok || value != 0 {
			t.Fatalf("%s = %d, %t", key, value, ok)
		}
	}
}

func TestConvertPlanMapsNegativeOneToTypedUnlimited(t *testing.T) {
	result, err := ConvertPlan(ConversionRequest{Snapshot: PlanSnapshot{
		Name: "Unlimited", Slug: "unlimited", BillingInterval: "monthly",
		Currency:     "USD",
		LimitsJSON:   []byte(`{"submissions_per_month":-1,"max_users":-1,"max_forms":-1,"max_storage_mb":-1}`),
		FeaturesJSON: []byte(`{}`),
	}, Available: true, SnapshotAt: snapshotInstant()})
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range grantValues(result.Version) {
		if value.Kind() == entitlement.Integer && !value.Unlimited() {
			t.Fatalf("%s was not unlimited", key)
		}
	}
}

func grantValues(version offeringdomain.Version) map[entitlement.Key]entitlement.Value {
	values := map[entitlement.Key]entitlement.Value{}
	for _, grant := range version.Grants() {
		values[grant.Definition().Key()] = grant.Value()
	}
	return values
}

func snapshotInstant() time.Time {
	return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
}
