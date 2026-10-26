package menuhttp

import "testing"

func TestCreateMenuConfig(t *testing.T) {
	if err := exerciseMenu("create_menu_config", "catalog.menu.update",
		func(handler *Handler) Route { return handler.CreateMenuConfigRoute() }); err != nil {
		t.Fatal(err)
	}
}
