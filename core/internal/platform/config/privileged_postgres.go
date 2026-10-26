package config

import "time"

type PrivilegedPostgres struct {
	endpointKind   PostgresEndpointKind
	dsn            Secret
	connectTimeout time.Duration
	cleanupTimeout time.Duration
}

func loadPrivilegedPostgres(
	values environmentValues,
	environment Environment,
	dsnKey string,
	expectedPrincipal string,
) (PrivilegedPostgres, error) {
	endpointKind, err := loadPostgresEndpointKind(values)
	if err != nil {
		return PrivilegedPostgres{}, err
	}
	dsn, err := loadSecret(values, dsnKey, true, secretFileLimit)
	if err != nil {
		return PrivilegedPostgres{}, err
	}
	if err := validatePostgresDSN(dsnKey, dsn, environment); err != nil {
		return PrivilegedPostgres{}, err
	}
	principal, _, err := postgresDSNPrincipalAndDatabase(dsn)
	if err != nil || principal != expectedPrincipal {
		return PrivilegedPostgres{}, configError(dsnKey, "uses an unexpected database principal")
	}
	connectTimeout, err := parseDuration(
		values, "HARPY_POSTGRES_CONNECT_TIMEOUT", "5s", time.Second, 30*time.Second,
	)
	if err != nil {
		return PrivilegedPostgres{}, err
	}
	cleanupTimeout, err := loadPostgresCleanupTimeout(values)
	if err != nil {
		return PrivilegedPostgres{}, err
	}
	return PrivilegedPostgres{
		endpointKind: endpointKind, dsn: dsn,
		connectTimeout: connectTimeout, cleanupTimeout: cleanupTimeout,
	}, nil
}

func (postgres PrivilegedPostgres) EndpointKind() PostgresEndpointKind {
	return postgres.endpointKind
}

func (postgres PrivilegedPostgres) DSN() Secret                   { return postgres.dsn }
func (postgres PrivilegedPostgres) ConnectTimeout() time.Duration { return postgres.connectTimeout }
func (postgres PrivilegedPostgres) CleanupTimeout() time.Duration { return postgres.cleanupTimeout }
