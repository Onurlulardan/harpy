package config

import "os"

type DatabaseRestore struct {
	runtime   DatabaseRestoreRuntime
	artifacts DatabaseRestoreArtifacts
}

type DatabaseRestoreRuntime struct {
	control        DedicatedRuntime
	scratch        Postgres
	scratchDSNFile ArtifactPath
}

func LoadDatabaseRestore() (DatabaseRestore, error) {
	return loadDatabaseRestore(os.Environ())
}

func loadDatabaseRestore(entries []string) (DatabaseRestore, error) {
	values, runtimeConfig, err := loadDedicatedRuntime(entries,
		profileDatabaseRestore,
		"HARPY_RESTORE_CONTROL_POSTGRES_DSN_FILE", "harpy_restore")
	if err != nil {
		return DatabaseRestore{}, err
	}
	scratchPath, err := loadReadArtifact(
		values, "HARPY_RESTORE_SCRATCH_POSTGRES_DSN_FILE", true,
	)
	if err != nil {
		return DatabaseRestore{}, err
	}
	scratch, err := loadPostgresFile(
		values, runtimeConfig.Environment(),
		"HARPY_RESTORE_SCRATCH_POSTGRES_DSN_FILE",
	)
	if err != nil {
		return DatabaseRestore{}, err
	}
	principal, scratchDatabase, err := postgresDSNPrincipalAndDatabase(scratch.DSN())
	_, controlDatabase, controlErr := postgresDSNPrincipalAndDatabase(
		runtimeConfig.Postgres().DSN(),
	)
	if err != nil || controlErr != nil || principal != "harpy_restore" ||
		scratchDatabase == controlDatabase {
		return DatabaseRestore{}, configError(
			"HARPY_RESTORE_SCRATCH_POSTGRES_DSN_FILE",
			"must use the restore principal and a distinct scratch database",
		)
	}
	runtime := DatabaseRestoreRuntime{
		control: runtimeConfig, scratch: scratch, scratchDSNFile: scratchPath,
	}
	artifacts, err := loadDatabaseRestoreArtifacts(values,
		runtimeConfig.DSNFile(), scratchPath)
	if err != nil {
		return DatabaseRestore{}, err
	}
	return DatabaseRestore{runtime: runtime, artifacts: artifacts}, nil
}

func (config DatabaseRestore) Runtime() DatabaseRestoreRuntime {
	return config.runtime
}

func (config DatabaseRestore) Artifacts() DatabaseRestoreArtifacts {
	return config.artifacts
}

func (config DatabaseRestoreRuntime) Environment() Environment {
	return config.control.Environment()
}
func (config DatabaseRestoreRuntime) ControlPostgres() Postgres {
	return config.control.Postgres()
}
func (config DatabaseRestoreRuntime) ScratchPostgres() Postgres {
	return config.scratch
}
func (config DatabaseRestoreRuntime) ControlDSNFile() ArtifactPath {
	return config.control.DSNFile()
}
func (config DatabaseRestoreRuntime) ScratchDSNFile() ArtifactPath {
	return config.scratchDSNFile
}
func (config DatabaseRestoreRuntime) LogLevel() LogLevel {
	return config.control.LogLevel()
}
func (config DatabaseRestoreRuntime) Telemetry() Telemetry {
	return config.control.Telemetry()
}
func (config DatabaseRestoreRuntime) Release() ReleaseInputs {
	return config.control.Release()
}
