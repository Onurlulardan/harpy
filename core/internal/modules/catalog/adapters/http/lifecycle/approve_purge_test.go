package lifecyclehttp

import "testing"

func TestApprovePurgeRouteCarriesExactApprovalAuthority(t *testing.T) {
	if err := exerciseInstallation(
		"approve_purge", "catalog.installation.purge.approve",
		func(handler *Handler) Route { return handler.ApprovePurgeRoute() },
	); err != nil {
		t.Fatal(err)
	}
}
