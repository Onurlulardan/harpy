package menuhttp

import "testing"

func TestReplaceMenu(t *testing.T) {
	if err := exerciseMenu("replace_menu", "catalog.menu.update",
		func(handler *Handler) Route { return handler.ReplaceMenuRoute() }); err != nil {
		t.Fatal(err)
	}
}
