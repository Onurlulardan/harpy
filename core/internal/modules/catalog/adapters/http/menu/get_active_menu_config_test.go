package menuhttp

import "testing"

func TestGetActiveMenuConfig(t *testing.T) {
	if err := exerciseMenu("get_active_menu_config", "catalog.menu.get",
		func(handler *Handler) Route { return handler.GetActiveMenuConfigRoute() }); err != nil {
		t.Fatal(err)
	}
}
