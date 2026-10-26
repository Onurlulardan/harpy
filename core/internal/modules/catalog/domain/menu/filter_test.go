package menu

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type targetAuthorizer struct{ allowed map[string]bool }

func (value targetAuthorizer) Allows(_ packagecontract.MenuKind, id packagecontract.ComponentID) bool {
	return value.allowed[id.String()]
}

func TestOverridesAndAuthorizationProduceEffectiveMenu(t *testing.T) {
	packageID, _ := packagecontract.ParsePackageID("018f2000-0000-7000-8000-000000000001")
	firstID, _ := packagecontract.ParseMenuItemID("018f2000-0000-7000-8000-000000000002")
	secondID, _ := packagecontract.ParseMenuItemID("018f2000-0000-7000-8000-000000000003")
	firstTarget, _ := packagecontract.ParseComponentID("018f2000-0000-7000-8000-000000000004")
	secondTarget, _ := packagecontract.ParseComponentID("018f2000-0000-7000-8000-000000000005")
	items := []Item{
		{MenuItem: packagecontract.MenuItem{ID: firstID, Kind: packagecontract.MenuForm,
			TargetID: &firstTarget, Label: "First", Order: 20}, PackageID: packageID, Visible: true},
		{MenuItem: packagecontract.MenuItem{ID: secondID, Kind: packagecontract.MenuDashboard,
			TargetID: &secondTarget, Label: "Second", Order: 10}, PackageID: packageID, Visible: true},
	}
	label, order, hidden := "Renamed", 1, false
	overridden, err := ApplyOverrides(items, []Override{
		{ItemID: firstID, Label: &label, Order: &order},
		{ItemID: secondID, Visible: &hidden},
	})
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Label != "First" || items[1].Visible != true {
		t.Fatal("override mutated source menu")
	}
	effective := Effective(overridden, map[string]bool{packageID.String(): true},
		targetAuthorizer{allowed: map[string]bool{firstTarget.String(): true}})
	if len(effective) != 1 || effective[0].ID.String() != firstID.String() ||
		effective[0].Label != "Renamed" || effective[0].Order != 1 {
		t.Fatalf("effective menu = %#v", effective)
	}
}

func TestEffectiveMenuFailsClosed(t *testing.T) {
	if result := Effective(nil, nil, nil); result != nil {
		t.Fatalf("nil authorizer returned %#v", result)
	}
}
