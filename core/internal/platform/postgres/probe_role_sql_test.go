package postgres

import (
	"strings"
	"testing"
)

func TestProbeRoleSQLSelectsFoundationRole(t *testing.T) {
	tests := []struct {
		profile Profile
		role    string
		sql     string
	}{
		{ProfileAPI, "harpy_api", probeAPIRoleSQL},
		{ProfileWorker, "harpy_worker", probeRuntimeRoleSQL},
		{ProfileOperatorInitialize, "harpy_api", probeAPIRoleSQL},
		{ProfileDatabaseRestore, "harpy_owner", probeOwnerRoleSQL},
	}
	for _, test := range tests {
		contract, err := test.profile.contract()
		if err != nil {
			t.Fatal(err)
		}
		sql := (&Pool{profile: test.profile, contract: contract}).probeRoleSQL()
		if sql != test.sql || !strings.Contains(sql, test.role) {
			t.Fatalf("%s probe = %q, want %s", test.profile, sql, test.role)
		}
		if strings.Contains(sql, "harpy_runtime") {
			t.Fatalf("%s probe sets harpy_runtime", test.profile)
		}
	}
}
