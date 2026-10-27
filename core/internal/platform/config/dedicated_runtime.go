package config

import "time"

type DedicatedRuntime struct {
	environment Environment
	postgres    Postgres
	dsnFile     ArtifactPath
	logLevel    LogLevel
	telemetry   Telemetry
	release     ReleaseInputs
}

func loadDedicatedRuntime(
	entries []string,
	current profile,
	dsnFileKey string,
	expectedPrincipal string,
) (environmentValues, DedicatedRuntime, error) {
	values, err := collectEnvironment(entries, current)
	if err != nil {
		return nil, DedicatedRuntime{}, err
	}
	rawEnvironment, found := values.lookup("HARPY_ENVIRONMENT")
	if !found {
		return nil, DedicatedRuntime{}, configError("HARPY_ENVIRONMENT", "is required")
	}
	environment, err := parseEnvironment(rawEnvironment)
	if err != nil {
		return nil, DedicatedRuntime{}, err
	}
	dsnPath, err := loadReadArtifact(values, dsnFileKey, true)
	if err != nil {
		return nil, DedicatedRuntime{}, err
	}
	postgres, err := loadPostgresFile(values, environment, dsnFileKey)
	if err != nil {
		return nil, DedicatedRuntime{}, err
	}
	principal, _, err := postgresDSNPrincipalAndDatabase(postgres.DSN())
	if err != nil || principal != expectedPrincipal {
		return nil, DedicatedRuntime{}, configError(dsnFileKey, "uses an unexpected database principal")
	}
	logLevel, err := parseLogLevel(values.valueOr("HARPY_LOG_LEVEL", "info"))
	if err != nil {
		return nil, DedicatedRuntime{}, err
	}
	telemetry, err := loadTelemetry(values, environment)
	if err != nil {
		return nil, DedicatedRuntime{}, err
	}
	release, err := loadReleaseInputs(values)
	if err != nil {
		return nil, DedicatedRuntime{}, err
	}
	return values, DedicatedRuntime{
		environment: environment, postgres: postgres, dsnFile: dsnPath,
		logLevel: logLevel, telemetry: telemetry, release: release,
	}, nil
}

func (config DedicatedRuntime) Environment() Environment { return config.environment }
func (config DedicatedRuntime) Postgres() Postgres       { return config.postgres }
func (config DedicatedRuntime) DSNFile() ArtifactPath    { return config.dsnFile }
func (config DedicatedRuntime) LogLevel() LogLevel       { return config.logLevel }
func (config DedicatedRuntime) Telemetry() Telemetry     { return config.telemetry }
func (config DedicatedRuntime) Release() ReleaseInputs   { return config.release }

type DedicatedLongRunning struct {
	runtime                 DedicatedRuntime
	managementListenAddress string
	drainTimeout            time.Duration
	readinessTimeout        time.Duration
}

func loadDedicatedLongRunning(
	values environmentValues,
	runtimeConfig DedicatedRuntime,
) (DedicatedLongRunning, error) {
	address, _, err := parseListenAddress(values,
		"HARPY_MANAGEMENT_LISTEN_ADDR", defaultManagementListenAddress)
	if err != nil {
		return DedicatedLongRunning{}, err
	}
	drain, err := parseDuration(values,
		"HARPY_DRAIN_TIMEOUT", "30s", 5*time.Second, 5*time.Minute)
	if err != nil {
		return DedicatedLongRunning{}, err
	}
	if runtimeConfig.Telemetry().ExportTimeout() > drain {
		return DedicatedLongRunning{}, configError(
			"HARPY_OTLP_EXPORT_TIMEOUT", "must not exceed the drain timeout")
	}
	readiness, err := parseDuration(values,
		"HARPY_READINESS_TIMEOUT", "2s", 250*time.Millisecond, 10*time.Second)
	if err != nil {
		return DedicatedLongRunning{}, err
	}
	if readiness > runtimeConfig.Postgres().HealthCheckPeriod() {
		return DedicatedLongRunning{}, configError(
			"HARPY_READINESS_TIMEOUT", "must not exceed the health-check period")
	}
	return DedicatedLongRunning{
		runtime: runtimeConfig, managementListenAddress: address,
		drainTimeout: drain, readinessTimeout: readiness,
	}, nil
}

func (config DedicatedLongRunning) Runtime() DedicatedRuntime { return config.runtime }
func (config DedicatedLongRunning) ManagementListenAddress() string {
	return config.managementListenAddress
}
func (config DedicatedLongRunning) DrainTimeout() time.Duration { return config.drainTimeout }
func (config DedicatedLongRunning) ReadinessTimeout() time.Duration {
	return config.readinessTimeout
}
