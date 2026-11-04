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
	roleInherit     bool
	connLimit       int
}

func (profile Profile) contract() (profileContract, error) {
	switch profile {
	case ProfileAPI:
		return applicationProfile(foundationRoleAPI, "core-api"), nil
	case ProfileWorker:
		return applicationProfile(foundationRoleWorker, "core-worker"), nil
	case ProfileOperatorInitialize:
		return runtimeProfile("core-operator-initialize"), nil
	case ProfileOperatorIdentityReconcile:
		return runtimeProfile("core-operator-identity-reconcile"), nil
	case ProfileOperatorLicenseReconcile:
		return runtimeProfile("core-operator-license-reconcile"), nil
	case ProfilePostRestoreObserver:
		return postRestoreObserverProfile(), nil
	case ProfileDatabaseBackupExecutor:
		return backupProfile(), nil
	case ProfileDatabaseRestore:
		return restoreProfile(), nil
	default:
		return profileContract{}, fmt.Errorf("postgres profile is invalid")
	}
}

const (
	foundationRoleAPI    = "harpy_api"
	foundationRoleWorker = "harpy_worker"
)

func applicationProfile(role, application string) profileContract {
	return profileContract{
		sessionUser: role, currentUser: role,
		applicationName: "harpy-v2:" + application, runtime: true,
		foundation: true, profileTx: true, currentLogin: true,
		roleInherit: true, connLimit: 20,
	}
}

func runtimeProfile(application string) profileContract {
	return profileContract{
		sessionUser: "harpy_v2_runtime", currentUser: "harpy_runtime",
		applicationName: "harpy-v2:" + application, runtime: true,
		foundation: true, profileTx: true, memberships: 1, memberSet: true,
		connLimit: -1,
	}
}

func backupProfile() profileContract {
	return profileContract{
		sessionUser: "harpy_backup", currentUser: "harpy_backup_reader",
		applicationName: "harpy-v2:core-database-backup-executor",
		foundation:      true, profileTx: true, memberships: 1, memberSet: true,
		connLimit: -1,
	}
}

func restoreProfile() profileContract {
	return profileContract{
		sessionUser: "harpy_restore", currentUser: "harpy_restore",
		applicationName: "harpy-v2:core-database-restore",
		profileTx:       true,
		currentLogin:    true, currentTemp: true, currentCreate: true,
		connLimit: -1,
	}
}

func postRestoreObserverProfile() profileContract {
	return profileContract{
		sessionUser:     "harpy_reactivation_observer",
		currentUser:     "harpy_reactivation_observer",
		applicationName: "harpy-v2:core-operator-post-restore-reactivation-observer",
		profileTx:       true,
		currentLogin:    true,
		connLimit:       -1,
	}
}
