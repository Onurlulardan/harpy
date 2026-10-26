package config

import "os"

const (
	cutoverDecoderPrincipal = "harpy_cutover_decoder"
	cutoverAdopterPrincipal = "harpy_adopter"
)

type CutoverAdoption struct {
	environment    Environment
	sourcePostgres PrivilegedPostgres
	targetPostgres PrivilegedPostgres
	artifacts      CutoverAdoptionArtifacts
	logLevel       LogLevel
	telemetry      Telemetry
}

func LoadCutoverAdoption() (CutoverAdoption, error) {
	return loadCutoverAdoption(os.Environ())
}

func loadCutoverAdoption(entries []string) (CutoverAdoption, error) {
	values, err := collectEnvironment(entries, profileCutoverAdoption)
	if err != nil {
		return CutoverAdoption{}, err
	}
	rawEnvironment, found := values.lookup("HARPY_ENVIRONMENT")
	if !found {
		return CutoverAdoption{}, configError("HARPY_ENVIRONMENT", "is required")
	}
	environment, err := parseEnvironment(rawEnvironment)
	if err != nil {
		return CutoverAdoption{}, err
	}
	source, err := loadPrivilegedPostgres(
		values, environment, "HARPY_CUTOVER_SOURCE_DSN", cutoverDecoderPrincipal,
	)
	if err != nil {
		return CutoverAdoption{}, err
	}
	target, err := loadPrivilegedPostgres(
		values, environment, "HARPY_CUTOVER_TARGET_DSN", cutoverAdopterPrincipal,
	)
	if err != nil {
		return CutoverAdoption{}, err
	}
	artifacts, err := loadCutoverAdoptionArtifacts(values)
	if err != nil {
		return CutoverAdoption{}, err
	}
	logLevel, err := parseLogLevel(values.valueOr("HARPY_LOG_LEVEL", "info"))
	if err != nil {
		return CutoverAdoption{}, err
	}
	telemetry, err := loadTelemetry(values, environment)
	if err != nil {
		return CutoverAdoption{}, err
	}
	return CutoverAdoption{
		environment: environment, sourcePostgres: source, targetPostgres: target,
		artifacts: artifacts, logLevel: logLevel, telemetry: telemetry,
	}, nil
}

func (c CutoverAdoption) Environment() Environment            { return c.environment }
func (c CutoverAdoption) SourcePostgres() PrivilegedPostgres  { return c.sourcePostgres }
func (c CutoverAdoption) TargetPostgres() PrivilegedPostgres  { return c.targetPostgres }
func (c CutoverAdoption) Artifacts() CutoverAdoptionArtifacts { return c.artifacts }
func (c CutoverAdoption) LogLevel() LogLevel                  { return c.logLevel }
func (c CutoverAdoption) Telemetry() Telemetry                { return c.telemetry }
