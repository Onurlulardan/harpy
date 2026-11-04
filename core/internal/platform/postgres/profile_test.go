package postgres

import "testing"

func TestProfilesPinRolesAndApplicationNames(t *testing.T) {
	tests := []struct {
		profile Profile
		session string
		current string
		runtime bool
	}{
		{ProfileAPI, "harpy_api", "harpy_api", true},
		{ProfileWorker, "harpy_worker", "harpy_worker", true},
		{ProfileOperatorInitialize, "harpy_api", "harpy_api", true},
		{ProfileOperatorIdentityReconcile, "harpy_api", "harpy_api", true},
		{ProfileOperatorLicenseReconcile, "harpy_api", "harpy_api", true},
		{ProfilePostRestoreObserver, "harpy_reactivation_observer",
			"harpy_reactivation_observer", false},
		{ProfileDatabaseBackupExecutor, "harpy_api", "harpy_api", false},
		{ProfileDatabaseRestore, "harpy_restore", "harpy_restore", false},
	}
	for _, test := range tests {
		contract, err := test.profile.contract()
		if err != nil {
			t.Fatal(err)
		}
		if contract.sessionUser != test.session || contract.currentUser != test.current ||
			contract.runtime != test.runtime || contract.applicationName == "" {
			t.Fatalf("unexpected profile contract: %+v", contract)
		}
	}
	for _, profile := range []Profile{
		ProfileOperatorInitialize, ProfileOperatorIdentityReconcile,
		ProfileOperatorLicenseReconcile, ProfileDatabaseBackupExecutor,
	} {
		contract, err := profile.contract()
		if err != nil || contract.memberships != 0 || contract.memberSet ||
			contract.sessionUser != foundationRoleAPI ||
			contract.currentUser != foundationRoleAPI || !contract.currentLogin ||
			!contract.roleInherit || contract.connLimit != 20 {
			t.Fatalf("%s does not run as harpy_api", profile)
		}
	}
	restore, err := ProfileDatabaseRestore.contract()
	if err != nil || restore.sessionUser != "harpy_restore" ||
		restore.currentUser != "harpy_restore" {
		t.Fatal("restore profile changed")
	}
	if _, err := Profile("admin").contract(); err == nil {
		t.Fatal("unknown PostgreSQL profile accepted")
	}
}
