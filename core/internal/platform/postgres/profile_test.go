package postgres

import "testing"

func TestProfilesPinRolesAndApplicationNames(t *testing.T) {
	tests := []struct {
		profile Profile
		session string
		current string
		runtime bool
	}{
		{ProfileAPI, "harpy_v2_runtime", "harpy_runtime", true},
		{ProfileWorker, "harpy_v2_runtime", "harpy_runtime", true},
		{ProfileOperatorInitialize, "harpy_v2_runtime", "harpy_runtime", true},
		{ProfileOperatorIdentityReconcile, "harpy_v2_runtime", "harpy_runtime", true},
		{ProfileOperatorLicenseReconcile, "harpy_v2_runtime", "harpy_runtime", true},
		{ProfilePostRestoreObserver, "harpy_reactivation_observer",
			"harpy_reactivation_observer", false},
		{ProfileMigratorDeploy, "harpy_migrator", "harpy_owner", false},
		{ProfileSchemaExecutor, "harpy_migrator", "harpy_owner", false},
		{ProfileTenantMigrationExecutor, "harpy_migrator", "harpy_owner", false},
		{ProfileDatabaseBackupExecutor, "harpy_backup", "harpy_backup_reader", false},
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
	if _, err := Profile("admin").contract(); err == nil {
		t.Fatal("unknown PostgreSQL profile accepted")
	}
}
