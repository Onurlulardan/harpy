package postgres

import "fmt"

type Profile string

const (
	ProfileAPI                       Profile = "api"
	ProfileWorker                    Profile = "worker"
	ProfileOperatorInitialize        Profile = "operator_initialize"
	ProfileOperatorIdentityReconcile Profile = "operator_identity_reconcile"
	ProfileOperatorLicenseReconcile  Profile = "operator_license_reconcile"
	ProfilePostRestoreObserver       Profile = "operator_post_restore_reactivation_observer"
	ProfileMigratorDeploy            Profile = "migrator_deploy"
	ProfileSchemaExecutor            Profile = "schema_executor"
	ProfileTenantMigrationExecutor   Profile = "tenant_migration_executor"
	ProfileDatabaseBackupExecutor    Profile = "database_backup_executor"
	ProfileDatabaseRestore           Profile = "database_restore"
)

type profileContract struct {
	sessionUser     string
	currentUser     string
	applicationName string
	runtime         bool
	foundation      bool
	profileTx       bool
	memberships     int
	memberSet       bool
	currentLogin    bool
	currentTemp     bool
	currentCreate   bool
}

func (profile Profile) contract() (profileContract, error) {
	switch profile {
	case ProfileAPI:
		return runtimeProfile("core-api"), nil
	case ProfileWorker:
		return runtimeProfile("core-worker"), nil
	case ProfileOperatorInitialize:
		return runtimeProfile("core-operator-initialize"), nil
	case ProfileOperatorIdentityReconcile:
		return runtimeProfile("core-operator-identity-reconcile"), nil
	case ProfileOperatorLicenseReconcile:
		return runtimeProfile("core-operator-license-reconcile"), nil
	case ProfilePostRestoreObserver:
		return postRestoreObserverProfile(), nil
	case ProfileMigratorDeploy:
		return migratorProfile("core-migrator-deploy"), nil
	case ProfileSchemaExecutor:
		return migratorProfile("core-schema-executor"), nil
	case ProfileTenantMigrationExecutor:
		return migratorProfile("core-tenant-migration-executor"), nil
	case ProfileDatabaseBackupExecutor:
		return backupProfile(), nil
	case ProfileDatabaseRestore:
		return restoreProfile(), nil
	default:
		return profileContract{}, fmt.Errorf("postgres profile is invalid")
	}
}

func runtimeProfile(application string) profileContract {
	return profileContract{
		sessionUser: "harpy_v2_runtime", currentUser: "harpy_runtime",
		applicationName: "harpy-v2:" + application, runtime: true,
		foundation: true, profileTx: true, memberships: 1, memberSet: true,
	}
}

func migratorProfile(application string) profileContract {
	return profileContract{
		sessionUser: "harpy_migrator", currentUser: "harpy_owner",
		applicationName: "harpy-v2:" + application,
		memberships:     1, memberSet: true, currentTemp: true, currentCreate: true,
	}
}

func backupProfile() profileContract {
	return profileContract{
		sessionUser: "harpy_backup", currentUser: "harpy_backup_reader",
		applicationName: "harpy-v2:core-database-backup-executor",
		foundation:      true, profileTx: true, memberships: 1, memberSet: true,
	}
}

func restoreProfile() profileContract {
	return profileContract{
		sessionUser: "harpy_restore", currentUser: "harpy_restore",
		applicationName: "harpy-v2:core-database-restore",
		profileTx:       true,
		currentLogin:    true, currentTemp: true, currentCreate: true,
	}
}

func postRestoreObserverProfile() profileContract {
	return profileContract{
		sessionUser:     "harpy_reactivation_observer",
		currentUser:     "harpy_reactivation_observer",
		applicationName: "harpy-v2:core-operator-post-restore-reactivation-observer",
		profileTx:       true,
		currentLogin:    true,
	}
}
