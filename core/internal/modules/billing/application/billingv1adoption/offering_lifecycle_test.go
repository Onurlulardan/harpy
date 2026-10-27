package billingv1adoptionapp

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestConvertOfferingSeparatesAvailabilityFromRetirement(t *testing.T) {
	created := time.Date(2026, 7, 17, 10, 0, 0, 123000, time.UTC)
	deleted := created.Add(time.Hour)
	tests := []struct {
		name      string
		active    bool
		deletedAt *time.Time
		available bool
		retired   bool
	}{
		{name: "active", active: true, available: true},
		{name: "inactive", active: false, available: false},
		{name: "deleted", active: true, deletedAt: &deleted, available: false, retired: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := completeSourcePlan(created)
			source.IsActive, source.DeletedAt = test.active, test.deletedAt
			record, _, err := convertOffering(source, false)
			if err != nil {
				t.Fatal(err)
			}
			if record.Available != test.available || (record.RetiredAt != nil) != test.retired {
				t.Fatalf("available/retired = %t/%t", record.Available, record.RetiredAt != nil)
			}
			if record.SourceID != source.ID.String() {
				t.Fatalf("source id = %q", record.SourceID)
			}
			canonical, parseErr := uuid.Parse(record.ID)
			if parseErr != nil || canonical.Version() != uuid.Version(7) || canonical == source.ID {
				t.Fatalf("canonical id = %q (%v)", record.ID, parseErr)
			}
		})
	}
}

func completeSourcePlan(created time.Time) SourcePlan {
	return SourcePlan{
		ID:   uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		Name: "Pro", Slug: "pro", Description: "Pro plan",
		PriceCents: 9900, BillingInterval: "monthly", Currency: "USD",
		LimitsJSON:   []byte(`{"submissions_per_month":500,"max_users":3,"max_forms":10,"max_storage_mb":100}`),
		FeaturesJSON: []byte(`{"ai_chat":true,"workflows":false,"api_access":true}`),
		SortOrder:    2, IsActive: true, IsPublic: true, TrialDays: 14,
		CreatedAt: &created,
	}
}
