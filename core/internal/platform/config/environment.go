package config

import "fmt"

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentTest        Environment = "test"
)

func parseEnvironment(value string) (Environment, error) {
	environment := Environment(value)
	switch environment {
	case EnvironmentDevelopment, EnvironmentProduction, EnvironmentStaging, EnvironmentTest:
		return environment, nil
	default:
		return "", configError("HARPY_ENVIRONMENT", "must be a supported environment")
	}
}

func (e Environment) String() string {
	return string(e)
}

func (e Environment) productionLike() bool {
	return e == EnvironmentProduction || e == EnvironmentStaging
}

type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogError LogLevel = "error"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
)

func parseLogLevel(value string) (LogLevel, error) {
	level := LogLevel(value)
	switch level {
	case LogDebug, LogError, LogInfo, LogWarn:
		return level, nil
	default:
		return "", configError("HARPY_LOG_LEVEL", "must be a supported level")
	}
}

func (l LogLevel) String() string {
	return string(l)
}

func configError(key, problem string) error {
	return fmt.Errorf("config %s: %s", key, problem)
}
