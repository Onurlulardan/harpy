package config

import "time"

const defaultManagementListenAddress = "0.0.0.0:9090"

type Runtime struct {
	environment      Environment
	postgres         Postgres
	readinessTimeout time.Duration
	logLevel         LogLevel
	telemetry        Telemetry
}

func loadRuntime(entries []string, current profile) (environmentValues, Runtime, error) {
	values, err := collectEnvironment(entries, current)
	if err != nil {
		return nil, Runtime{}, err
	}
	rawEnvironment, found := values.lookup("HARPY_ENVIRONMENT")
	if !found {
		return nil, Runtime{}, configError("HARPY_ENVIRONMENT", "is required")
	}
	environment, err := parseEnvironment(rawEnvironment)
	if err != nil {
		return nil, Runtime{}, err
	}
	postgres, err := loadPostgres(values, environment)
	if err != nil {
		return nil, Runtime{}, err
	}
	readinessTimeout, err := parseDuration(values, "HARPY_READINESS_TIMEOUT", "2s", 250*time.Millisecond, 10*time.Second)
	if err != nil {
		return nil, Runtime{}, err
	}
	if readinessTimeout > postgres.HealthCheckPeriod() {
		return nil, Runtime{}, configError("HARPY_READINESS_TIMEOUT", "must not exceed the health-check period")
	}
	logLevel, err := parseLogLevel(values.valueOr("HARPY_LOG_LEVEL", "info"))
	if err != nil {
		return nil, Runtime{}, err
	}
	telemetry, err := loadTelemetry(values, environment)
	if err != nil {
		return nil, Runtime{}, err
	}
	return values, Runtime{
		environment: environment, postgres: postgres, readinessTimeout: readinessTimeout,
		logLevel: logLevel, telemetry: telemetry,
	}, nil
}

func (r Runtime) Environment() Environment        { return r.environment }
func (r Runtime) Postgres() Postgres              { return r.postgres }
func (r Runtime) ReadinessTimeout() time.Duration { return r.readinessTimeout }
func (r Runtime) LogLevel() LogLevel              { return r.logLevel }
func (r Runtime) Telemetry() Telemetry            { return r.telemetry }

type LongRunning struct {
	runtime                 Runtime
	managementListenAddress string
	drainTimeout            time.Duration
}

func loadLongRunning(values environmentValues, runtimeConfig Runtime) (LongRunning, error) {
	address, _, err := parseListenAddress(values, "HARPY_MANAGEMENT_LISTEN_ADDR", defaultManagementListenAddress)
	if err != nil {
		return LongRunning{}, err
	}
	drainTimeout, err := parseDuration(values, "HARPY_DRAIN_TIMEOUT", "30s", 5*time.Second, 5*time.Minute)
	if err != nil {
		return LongRunning{}, err
	}
	if runtimeConfig.Telemetry().ExportTimeout() > drainTimeout {
		return LongRunning{}, configError(
			"HARPY_OTLP_EXPORT_TIMEOUT", "must not exceed the drain timeout",
		)
	}
	return LongRunning{runtime: runtimeConfig, managementListenAddress: address, drainTimeout: drainTimeout}, nil
}

func (c LongRunning) Runtime() Runtime                { return c.runtime }
func (c LongRunning) ManagementListenAddress() string { return c.managementListenAddress }
func (c LongRunning) DrainTimeout() time.Duration     { return c.drainTimeout }
