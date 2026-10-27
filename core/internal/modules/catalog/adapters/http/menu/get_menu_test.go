package menuhttp

import "testing"

func TestGetMenu(t *testing.T) {
	if err := exerciseMenu("get_menu", "catalog.menu.get",
		func(handler *Handler) Route { return handler.GetMenuRoute() }); err != nil {
		t.Fatal(err)
	}
}
