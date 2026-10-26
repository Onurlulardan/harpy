package config

import "os"

type MigratorDeploy struct {
	runtime      Runtime
	verification ReleaseVerification
	artifacts    MigrationArtifacts
	catalogKeys  CatalogPackageKeyring
}

func LoadMigratorDeploy() (MigratorDeploy, error) {
	return loadMigratorDeploy(os.Environ())
}

func loadMigratorDeploy(entries []string) (MigratorDeploy, error) {
	values, runtimeConfig, err := loadRuntime(entries, profileMigratorDeploy)
	if err != nil {
		return MigratorDeploy{}, err
	}
	verification, err := loadReleaseVerification(values)
	if err != nil {
		return MigratorDeploy{}, err
	}
	artifacts, err := loadMigrationArtifacts(values)
	if err != nil {
		return MigratorDeploy{}, err
	}
	catalogKeys, err := loadCatalogPackageKeyring(values)
	if err != nil {
		return MigratorDeploy{}, err
	}
	if err := validateArtifactPathIsolation(verification.paths(), artifacts.paths()); err != nil {
		return MigratorDeploy{}, err
	}
	return MigratorDeploy{
		runtime: runtimeConfig, verification: verification, artifacts: artifacts,
		catalogKeys: catalogKeys,
	}, nil
}

func (c MigratorDeploy) Verification() ReleaseVerification { return c.verification }
func (c MigratorDeploy) Artifacts() MigrationArtifacts     { return c.artifacts }
func (c MigratorDeploy) CatalogPackageKeyring() CatalogPackageKeyring {
	return c.catalogKeys
}

func (c MigratorDeploy) Runtime() Runtime {
	return c.runtime
}

type SchemaExecutor struct {
	longRunning LongRunning
}

func LoadSchemaExecutor() (SchemaExecutor, error) {
	return loadSchemaExecutor(os.Environ())
}

func loadSchemaExecutor(entries []string) (SchemaExecutor, error) {
	values, runtimeConfig, err := loadRuntime(entries, profileSchemaExecutor)
	if err != nil {
		return SchemaExecutor{}, err
	}
	longRunning, err := loadLongRunning(values, runtimeConfig)
	if err != nil {
		return SchemaExecutor{}, err
	}
	return SchemaExecutor{longRunning: longRunning}, nil
}

func (c SchemaExecutor) LongRunning() LongRunning {
	return c.longRunning
}
