package lifecyclehttp

import "testing"

func TestRetire(t *testing.T) {
	if err := exerciseInstallation("retire", "catalog.installation.retire",
		func(handler *Handler) Route { return handler.RetireRoute() }); err != nil {
		t.Fatal(err)
	}
}
