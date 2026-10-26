package config

import "os"

type ClusterBootstrap struct {
	environment Environment
	postgres    BootstrapPostgres
	release     ReleaseAttestation
	artifacts   BootstrapArtifacts
	logLevel    LogLevel
}

func LoadClusterBootstrap() (ClusterBootstrap, error) {
	return loadClusterBootstrap(os.Environ())
}

func loadClusterBootstrap(entries []string) (ClusterBootstrap, error) {
	values, err := collectEnvironment(entries, profileClusterBootstrap)
	if err != nil {
		return ClusterBootstrap{}, err
	}
	rawEnvironment, found := values.lookup("HARPY_ENVIRONMENT")
	if !found {
		return ClusterBootstrap{}, configError("HARPY_ENVIRONMENT", "is required")
	}
	environment, err := parseEnvironment(rawEnvironment)
	if err != nil {
		return ClusterBootstrap{}, err
	}
	postgres, err := loadBootstrapPostgres(values, environment)
	if err != nil {
		return ClusterBootstrap{}, err
	}
	release, err := loadReleaseAttestation(values, "HARPY_BOOTSTRAP_RECEIPT_PATH")
	if err != nil {
		return ClusterBootstrap{}, err
	}
	artifacts, err := loadBootstrapArtifacts(values)
	if err != nil {
		return ClusterBootstrap{}, err
	}
	if err := validateArtifactPathIsolation(release.paths(), artifacts.paths()); err != nil {
		return ClusterBootstrap{}, err
	}
	logLevel, err := parseLogLevel(values.valueOr("HARPY_LOG_LEVEL", "info"))
	if err != nil {
		return ClusterBootstrap{}, err
	}
	return ClusterBootstrap{
		environment: environment, postgres: postgres, release: release,
		artifacts: artifacts, logLevel: logLevel,
	}, nil
}

func (config ClusterBootstrap) Environment() Environment      { return config.environment }
func (config ClusterBootstrap) Postgres() BootstrapPostgres   { return config.postgres }
func (config ClusterBootstrap) Release() ReleaseAttestation   { return config.release }
func (config ClusterBootstrap) Artifacts() BootstrapArtifacts { return config.artifacts }
func (config ClusterBootstrap) LogLevel() LogLevel            { return config.logLevel }
