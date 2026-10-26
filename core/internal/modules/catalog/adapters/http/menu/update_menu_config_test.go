package menuhttp

import "testing"

func TestUpdateMenuConfig(t *testing.T) {
	if err := exerciseMenu("update_menu_config", "catalog.menu.update",
		func(handler *Handler) Route { return handler.UpdateMenuConfigRoute() }); err != nil {
		t.Fatal(err)
	}
}
