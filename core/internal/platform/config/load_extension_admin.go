package config

import "os"

type ExtensionAdmin struct {
	environment Environment
	postgres    PrivilegedPostgres
	release     ReleaseAttestation
	artifacts   ExtensionArtifacts
	logLevel    LogLevel
}

func LoadExtensionAdmin() (ExtensionAdmin, error) {
	return loadExtensionAdmin(os.Environ())
}

func loadExtensionAdmin(entries []string) (ExtensionAdmin, error) {
	values, err := collectEnvironment(entries, profileExtensionAdmin)
	if err != nil {
		return ExtensionAdmin{}, err
	}
	environment, postgres, release, logLevel, err := loadPrivilegedProfile(
		values, "HARPY_EXTENSION_ADMIN_POSTGRES_DSN", "harpy_extension_admin",
		"HARPY_EXTENSION_RECEIPT_PATH",
	)
	if err != nil {
		return ExtensionAdmin{}, err
	}
	artifacts, err := loadExtensionArtifacts(values)
	if err != nil {
		return ExtensionAdmin{}, err
	}
	if err := validateArtifactPathIsolation(release.paths(), artifacts.paths()); err != nil {
		return ExtensionAdmin{}, err
	}
	return ExtensionAdmin{
		environment: environment, postgres: postgres, release: release,
		artifacts: artifacts, logLevel: logLevel,
	}, nil
}

func (config ExtensionAdmin) Environment() Environment      { return config.environment }
func (config ExtensionAdmin) Postgres() PrivilegedPostgres  { return config.postgres }
func (config ExtensionAdmin) Release() ReleaseAttestation   { return config.release }
func (config ExtensionAdmin) Artifacts() ExtensionArtifacts { return config.artifacts }
func (config ExtensionAdmin) LogLevel() LogLevel            { return config.logLevel }
