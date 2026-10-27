package menuhttp

import "testing"

func TestUpdateMenuOverride(t *testing.T) {
	if err := exerciseMenu("update_menu_override", "catalog.menu.update",
		func(handler *Handler) Route { return handler.UpdateMenuOverrideRoute() }); err != nil {
		t.Fatal(err)
	}
}
