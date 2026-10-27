package billingv1adoptionapp

import "testing"

func TestConvertPlanRejectsAmbiguousDocuments(t *testing.T) {
	tests := []struct {
		name     string
		limits   string
		features string
	}{
		{"negative", `{"submissions_per_month":-2}`, `{}`},
		{"limit null", `{"submissions_per_month":null}`, `{}`},
		{"limit string", `{"max_users":"3"}`, `{}`},
		{"unknown limit", `{"custom":1}`, `{}`},
		{"feature null", `{}`, `{"workflows":null}`},
		{"feature number", `{}`, `{"workflows":1}`},
		{"unknown feature", `{}`, `{"custom":true}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ConvertPlan(ConversionRequest{Snapshot: PlanSnapshot{
				Name: "Plan", Slug: "plan", BillingInterval: "monthly", Currency: "USD",
				LimitsJSON: []byte(test.limits), FeaturesJSON: []byte(test.features),
			}, SnapshotAt: snapshotInstant()})
			if err == nil {
				t.Fatal("ambiguous document accepted")
			}
		})
	}
}

func TestConvertPlanRejectsUnsupportedMoneyAndInterval(t *testing.T) {
	tests := []PlanSnapshot{
		{Name: "JPY", Slug: "jpy", BillingInterval: "monthly", Currency: "JPY"},
		{Name: "Year", Slug: "year", BillingInterval: "yearly", Currency: "USD"},
		{Name: "Negative", Slug: "negative", PriceCents: -1,
			BillingInterval: "monthly", Currency: "USD"},
	}
	for _, snapshot := range tests {
		snapshot.LimitsJSON, snapshot.FeaturesJSON = []byte(`{}`), []byte(`{}`)
		if _, err := ConvertPlan(ConversionRequest{
			Snapshot: snapshot, SnapshotAt: snapshotInstant(),
		}); err == nil {
			t.Fatalf("invalid plan accepted: %#v", snapshot)
		}
	}
}

func TestStorageConversionRejectsOverflow(t *testing.T) {
	if _, err := storageValue(9_000_000_000); err == nil {
		t.Fatal("storage amount above policy bound accepted")
	}
}
