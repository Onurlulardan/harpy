package worker

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestAuthorityModesRouteToExactlyOnePurposeSeparatedWorkload(t *testing.T) {
	tests := map[config.WorkerAuthorityMode]authorityWorkloadRoute{
		config.WorkerAuthorityConnected:   authorityRouteLicenseRefresh,
		config.WorkerAuthorityAirGap:      authorityRouteLicenseRefresh,
		config.WorkerAuthorityDevelopment: authorityRouteLicenseRefresh,
		config.WorkerAuthoritySaaS:        authorityRouteSaaSRegistry,
	}
	for mode, wanted := range tests {
		got, err := routeAuthorityWorkload(mode)
		if err != nil || got != wanted {
			t.Fatalf("mode %s route=%s err=%v", mode, got, err)
		}
	}
	if _, err := routeAuthorityWorkload("combined"); err == nil {
		t.Fatal("combined authority route accepted")
	}
}
