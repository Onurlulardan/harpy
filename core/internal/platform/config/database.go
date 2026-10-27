package config

import "time"

type Postgres struct {
	endpointKind      PostgresEndpointKind
	dsn               Secret
	maxConnections    int32
	minConnections    int32
	connectTimeout    time.Duration
	cleanupTimeout    time.Duration
	maxLifetime       time.Duration
	maxIdleTime       time.Duration
	healthCheckPeriod time.Duration
}

func loadPostgres(values environmentValues, environment Environment) (Postgres, error) {
	dsn, err := loadSecret(values, "HARPY_POSTGRES_DSN", true, secretFileLimit)
	if err != nil {
		return Postgres{}, err
	}
	return loadPostgresSettings(values, environment, "HARPY_POSTGRES_DSN", dsn)
}

func loadPostgresFile(
	values environmentValues,
	environment Environment,
	dsnFileKey string,
) (Postgres, error) {
	path, found := values.lookup(dsnFileKey)
	if !found {
		return Postgres{}, configError(dsnFileKey, "is required")
	}
	raw, err := readSecretFile(dsnFileKey, path, secretFileLimit)
	if err != nil {
		return Postgres{}, err
	}
	if err := validateSecretValue(dsnFileKey, raw, secretFileLimit); err != nil {
		return Postgres{}, err
	}
	return loadPostgresSettings(values, environment, dsnFileKey, Secret{value: raw})
}

func loadPostgresSettings(
	values environmentValues,
	environment Environment,
	dsnKey string,
	dsn Secret,
) (Postgres, error) {
	endpointKind, err := loadPostgresEndpointKind(values)
	if err != nil {
		return Postgres{}, err
	}
	if err := validatePostgresDSN(dsnKey, dsn, environment); err != nil {
		return Postgres{}, err
	}
	maximum, err := parseInteger(values, "HARPY_POSTGRES_MAX_CONNS", "20", 2, 200)
	if err != nil {
		return Postgres{}, err
	}
	minimum, err := parseInteger(values, "HARPY_POSTGRES_MIN_CONNS", "2", 0, maximum)
	if err != nil {
		return Postgres{}, err
	}
	connectTimeout, err := parseDuration(values, "HARPY_POSTGRES_CONNECT_TIMEOUT", "5s", time.Second, 30*time.Second)
	if err != nil {
		return Postgres{}, err
	}
	cleanupTimeout, err := loadPostgresCleanupTimeout(values)
	if err != nil {
		return Postgres{}, err
	}
	maxLifetime, err := parseDuration(values, "HARPY_POSTGRES_MAX_CONN_LIFETIME", "30m", time.Minute, 24*time.Hour)
	if err != nil {
		return Postgres{}, err
	}
	maxIdleTime, err := parseDuration(values, "HARPY_POSTGRES_MAX_CONN_IDLE_TIME", "5m", 30*time.Second, time.Hour)
	if err != nil {
		return Postgres{}, err
	}
	if maxIdleTime > maxLifetime {
		return Postgres{}, configError("HARPY_POSTGRES_MAX_CONN_IDLE_TIME", "must not exceed connection lifetime")
	}
	healthPeriod, err := parseDuration(values, "HARPY_POSTGRES_HEALTH_CHECK_PERIOD", "30s", 5*time.Second, 5*time.Minute)
	if err != nil {
		return Postgres{}, err
	}
	return Postgres{
		endpointKind: endpointKind, dsn: dsn,
		maxConnections: int32(maximum), minConnections: int32(minimum),
		connectTimeout: connectTimeout, cleanupTimeout: cleanupTimeout, maxLifetime: maxLifetime,
		maxIdleTime: maxIdleTime, healthCheckPeriod: healthPeriod,
	}, nil
}

func (p Postgres) EndpointKind() PostgresEndpointKind   { return p.endpointKind }
func (p Postgres) DSN() Secret                          { return p.dsn }
func (p Postgres) MaxConnections() int32                { return p.maxConnections }
func (p Postgres) MinConnections() int32                { return p.minConnections }
func (p Postgres) ConnectTimeout() time.Duration        { return p.connectTimeout }
func (p Postgres) CleanupTimeout() time.Duration        { return p.cleanupTimeout }
func (p Postgres) MaxConnectionLifetime() time.Duration { return p.maxLifetime }
func (p Postgres) MaxConnectionIdleTime() time.Duration { return p.maxIdleTime }
func (p Postgres) HealthCheckPeriod() time.Duration     { return p.healthCheckPeriod }
