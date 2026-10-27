package billingv1

import (
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestMapPlanCreatePreservesInactiveAvailability(t *testing.T) {
	name, slug, inactive := "Hidden", "hidden", false
	command, err := mapPlanCreate(
		billingapi.Access{}, "create-hidden",
		planRequest{Name: &name, Slug: &slug, IsActive: &inactive},
		time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if command.InitialVersion.Available {
		t.Fatal("inactive V1 plan became available")
	}
}

func TestMapPlanUpdateCreatesImmutableAvailabilityVersions(t *testing.T) {
	current := billingapi.OfferingView{
		ID: "11111111-1111-4111-8111-111111111111", Key: "pro",
		Revision: 7, VersionID: "22222222-2222-4222-8222-222222222222",
		Name: "Pro", State: "active", Available: true,
		BillingInterval: "monthly",
		Prices: []billingapi.PriceView{{
			Cadence: "monthly",
			Amount:  billingapi.MoneyView{MinorUnits: 1000, Currency: "USD", RegistryVersion: 1},
		}},
	}
	for _, available := range []bool{false, true} {
		current.Available = !available
		command, err := mapPlanUpdate(
			billingapi.Access{}, "update-pro", current,
			planRequest{IsActive: &available},
			time.Date(2026, 7, 17, 11, 0, 0, 0, time.UTC),
		)
		if err != nil {
			t.Fatal(err)
		}
		if command.OfferingID != current.ID || command.ExpectedRevision != current.Revision ||
			command.Version.Available != available {
			t.Fatalf("command = %#v", command)
		}
	}
}
