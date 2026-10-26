package config

import "os"

type TenantMigrationExecutor struct {
	longRunning    LongRunning
	deploymentMode DeploymentMode
	release        ReleaseInputs
}

func LoadTenantMigrationExecutor() (TenantMigrationExecutor, error) {
	return loadTenantMigrationExecutor(os.Environ())
}

func loadTenantMigrationExecutor(entries []string) (TenantMigrationExecutor, error) {
	values, runtimeConfig, err := loadRuntime(entries, profileTenantMigrationExecutor)
	if err != nil {
		return TenantMigrationExecutor{}, err
	}
	longRunning, err := loadLongRunning(values, runtimeConfig)
	if err != nil {
		return TenantMigrationExecutor{}, err
	}
	mode, err := loadDeploymentMode(values)
	if err != nil {
		return TenantMigrationExecutor{}, err
	}
	release, err := loadReleaseInputs(values)
	if err != nil {
		return TenantMigrationExecutor{}, err
	}
	return TenantMigrationExecutor{
		longRunning: longRunning, deploymentMode: mode, release: release,
	}, nil
}

func (config TenantMigrationExecutor) LongRunning() LongRunning {
	return config.longRunning
}
func (config TenantMigrationExecutor) DeploymentMode() DeploymentMode {
	return config.deploymentMode
}
func (config TenantMigrationExecutor) Release() ReleaseInputs { return config.release }
