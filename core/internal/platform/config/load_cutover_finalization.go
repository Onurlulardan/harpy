package config

import "os"

const cutoverFinalizerPrincipal = "harpy_cutover_finalizer"

type CutoverFinalization struct {
	environment Environment
	postgres    PrivilegedPostgres
	artifacts   CutoverFinalizationArtifacts
	logLevel    LogLevel
	telemetry   Telemetry
}

func LoadCutoverFinalization() (CutoverFinalization, error) {
	return loadCutoverFinalization(os.Environ())
}

func loadCutoverFinalization(entries []string) (CutoverFinalization, error) {
	values, err := collectEnvironment(entries, profileCutoverFinalization)
	if err != nil {
		return CutoverFinalization{}, err
	}
	rawEnvironment, found := values.lookup("HARPY_ENVIRONMENT")
	if !found {
		return CutoverFinalization{}, configError("HARPY_ENVIRONMENT", "is required")
	}
	environment, err := parseEnvironment(rawEnvironment)
	if err != nil {
		return CutoverFinalization{}, err
	}
	postgres, err := loadPrivilegedPostgres(
		values, environment, "HARPY_CUTOVER_FINALIZER_DSN", cutoverFinalizerPrincipal,
	)
	if err != nil {
		return CutoverFinalization{}, err
	}
	artifacts, err := loadCutoverFinalizationArtifacts(values)
	if err != nil {
		return CutoverFinalization{}, err
	}
	logLevel, err := parseLogLevel(values.valueOr("HARPY_LOG_LEVEL", "info"))
	if err != nil {
		return CutoverFinalization{}, err
	}
	telemetry, err := loadTelemetry(values, environment)
	if err != nil {
		return CutoverFinalization{}, err
	}
	return CutoverFinalization{
		environment: environment, postgres: postgres, artifacts: artifacts,
		logLevel: logLevel, telemetry: telemetry,
	}, nil
}

func (c CutoverFinalization) Environment() Environment                { return c.environment }
func (c CutoverFinalization) Postgres() PrivilegedPostgres            { return c.postgres }
func (c CutoverFinalization) Artifacts() CutoverFinalizationArtifacts { return c.artifacts }
func (c CutoverFinalization) LogLevel() LogLevel                      { return c.logLevel }
func (c CutoverFinalization) Telemetry() Telemetry                    { return c.telemetry }
