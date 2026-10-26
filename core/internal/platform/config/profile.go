package config

import "strings"

type profile string

const (
	profileAPI                       profile = "api"
	profileClusterBootstrap          profile = "cluster_bootstrap"
	profileCutoverAdoption           profile = "cutover_adoption"
	profileCutoverFinalization       profile = "cutover_finalization"
	profileDatabaseBackupExecutor    profile = "database_backup_executor"
	profileDatabaseRestore           profile = "database_restore"
	profileExtensionAdmin            profile = "extension_admin"
	profileMigratorDeploy            profile = "migrator_deploy"
	profileOperatorIdentityReconcile profile = "operator_identity_reconcile"
	profileOperatorInitialize        profile = "operator_initialize"
	profileOperatorLicenseReconcile  profile = "operator_license_reconcile"
	profilePostRestoreObserver       profile = "operator_post_restore_reactivation_observer"
	profileSchemaExecutor            profile = "schema_executor"
	profileTenantMigrationExecutor   profile = "tenant_migration_executor"
	profileWorker                    profile = "worker"
)

type environmentValues map[string]string

func collectEnvironment(entries []string, current profile) (environmentValues, error) {
	allowed := allowedKeys(current)
	values := make(environmentValues)
	for _, entry := range entries {
		key, value, found := strings.Cut(entry, "=")
		managedWorkerKey := current == profileWorker && key == workerMemoryLimitKey
		if !found || !managedWorkerKey && key != deploymentModeEnvironmentKey &&
			!strings.HasPrefix(strings.ToUpper(key), "HARPY_") {
			continue
		}
		if !allowed[key] {
			return nil, configError("HARPY_*", "profile contains an unknown key")
		}
		if _, duplicate := values[key]; duplicate {
			return nil, configError(key, "is duplicated")
		}
		if value == "" {
			return nil, configError(key, "must not be empty")
		}
		values[key] = value
	}
	return values, nil
}

func (v environmentValues) lookup(key string) (string, bool) {
	value, exists := v[key]
	return value, exists
}

func (v environmentValues) valueOr(key, fallback string) string {
	if value, exists := v[key]; exists {
		return value
	}
	return fallback
}

func allowedKeys(current profile) map[string]bool {
	var groups [][]string
	switch current {
	case profileAPI:
		groups = [][]string{runtimeKeys(), longRunningKeys(), apiKeys(), apiPhase14Keys()}
	case profileWorker:
		groups = [][]string{runtimeKeys(), longRunningKeys(), workerKeys(), workerPhase14Keys()}
	case profileOperatorInitialize:
		groups = [][]string{runtimeKeys(), operatorInitializeKeys()}
	case profileOperatorIdentityReconcile:
		groups = [][]string{runtimeKeys(), operatorIdentityReconcileKeys()}
	case profileOperatorLicenseReconcile:
		groups = [][]string{runtimeKeys(), operatorLicenseReconcileKeys()}
	case profilePostRestoreObserver:
		groups = [][]string{postRestoreObserverKeys()}
	case profileTenantMigrationExecutor:
		groups = [][]string{
			runtimeKeys(), longRunningKeys(), releaseInputKeys(),
			{deploymentModeEnvironmentKey},
		}
	case profileDatabaseBackupExecutor:
		groups = [][]string{databaseBackupExecutorKeys()}
	case profileDatabaseRestore:
		groups = [][]string{databaseRestoreKeys()}
	case profileSchemaExecutor:
		groups = [][]string{runtimeKeys(), longRunningKeys()}
	case profileMigratorDeploy:
		groups = [][]string{runtimeKeys(), releaseVerificationKeys()}
	case profileClusterBootstrap:
		groups = [][]string{clusterBootstrapKeys()}
	case profileExtensionAdmin:
		groups = [][]string{
			privilegedKeys("HARPY_EXTENSION_ADMIN_POSTGRES_DSN", "HARPY_EXTENSION_RECEIPT_PATH"),
			extensionArtifactKeys(),
		}
	case profileCutoverAdoption:
		groups = [][]string{cutoverAdoptionKeys()}
	case profileCutoverFinalization:
		groups = [][]string{cutoverFinalizationKeys()}
	default:
		return map[string]bool{}
	}
	keys := make(map[string]bool)
	for _, group := range groups {
		for _, key := range group {
			keys[key] = true
		}
	}
	return keys
}
