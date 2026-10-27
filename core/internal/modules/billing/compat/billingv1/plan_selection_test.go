package billingv1

import (
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestSelectAdminPlansIncludesUnavailableAndSorts(t *testing.T) {
	selected, err := selectAdminPlans([]billingapi.OfferingView{
		planView("b", "beta", 20, true, true, "active"),
		planView("retired", "old", 0, true, true, "retired"),
		planView("a", "alpha", 10, false, false, "active"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].ID != "a" || selected[1].ID != "b" || selected[0].Available {
		t.Fatalf("selected = %#v", selected)
	}
}

func TestSelectPublicPlansRequiresLatestPublicAvailability(t *testing.T) {
	for _, offering := range []billingapi.OfferingView{
		planView("hidden", "hidden", 1, false, true, "active"),
		planView("private", "private", 1, true, false, "active"),
	} {
		if _, err := selectPublicPlans([]billingapi.OfferingView{offering}); !errors.Is(err, ErrSnapshotRequired) {
			t.Fatalf("offering %#v: error = %v", offering, err)
		}
	}
}

func TestSelectAdminPlansUsesLatestVersionPerRoot(t *testing.T) {
	older := planView("versioned", "versioned", 1, true, true, "active")
	older.Version = 1
	latest := older
	latest.Version = 2
	latest.VersionID = "versioned-version-2"
	latest.Available = false
	selected, err := selectAdminPlans([]billingapi.OfferingView{latest, older})
	if err != nil || len(selected) != 1 || selected[0].Version != 2 || selected[0].Available {
		t.Fatalf("selected = %#v, error = %v", selected, err)
	}
	if _, err := selectAdminPlans([]billingapi.OfferingView{latest, latest}); !errors.Is(err, ErrSnapshotRequired) {
		t.Fatalf("error = %v", err)
	}
}

func TestPlanWithLegacyIDRejectsWrongCanonicalKind(t *testing.T) {
	offering := planView("canonical", "pro", 1, true, true, "active")
	_, err := planWithLegacyID(offering, billingapi.LegacyIdentityView{
		LegacyID: "legacy", CanonicalID: "another-canonical",
	})
	if !errors.Is(err, ErrSnapshotRequired) {
		t.Fatalf("error = %v", err)
	}
}

func planView(id, key string, order int32, available, public bool, state string) billingapi.OfferingView {
	return billingapi.OfferingView{
		ID: id, VersionID: id + "-version", Key: key, SortOrder: order,
		Available: available, Public: public, State: state, Version: 1,
	}
}
