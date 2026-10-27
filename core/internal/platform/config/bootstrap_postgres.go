package config

import "time"

type BootstrapPostgres struct {
	endpointKind   PostgresEndpointKind
	maintenanceDSN Secret
	targetDSN      Secret
	connectTimeout time.Duration
	cleanupTimeout time.Duration
}

func loadBootstrapPostgres(
	values environmentValues,
	environment Environment,
) (BootstrapPostgres, error) {
	endpointKind, err := loadPostgresEndpointKind(values)
	if err != nil {
		return BootstrapPostgres{}, err
	}
	maintenance, err := loadSecret(
		values, "HARPY_BOOTSTRAP_MAINTENANCE_DSN", true, secretFileLimit,
	)
	if err != nil {
		return BootstrapPostgres{}, err
	}
	target, err := loadSecret(values, "HARPY_BOOTSTRAP_TARGET_DSN", true, secretFileLimit)
	if err != nil {
		return BootstrapPostgres{}, err
	}
	if err := validateBootstrapDSNs(environment, maintenance, target); err != nil {
		return BootstrapPostgres{}, err
	}
	connectTimeout, err := parseDuration(
		values, "HARPY_POSTGRES_CONNECT_TIMEOUT", "5s", time.Second, 30*time.Second,
	)
	if err != nil {
		return BootstrapPostgres{}, err
	}
	cleanupTimeout, err := loadPostgresCleanupTimeout(values)
	if err != nil {
		return BootstrapPostgres{}, err
	}
	return BootstrapPostgres{
		endpointKind: endpointKind, maintenanceDSN: maintenance, targetDSN: target,
		connectTimeout: connectTimeout, cleanupTimeout: cleanupTimeout,
	}, nil
}

func validateBootstrapDSNs(environment Environment, maintenance, target Secret) error {
	const maintenanceKey = "HARPY_BOOTSTRAP_MAINTENANCE_DSN"
	const targetKey = "HARPY_BOOTSTRAP_TARGET_DSN"
	if err := validatePostgresDSN(maintenanceKey, maintenance, environment); err != nil {
		return err
	}
	if err := validatePostgresDSN(targetKey, target, environment); err != nil {
		return err
	}
	maintenancePrincipal, database, err := postgresDSNPrincipalAndDatabase(maintenance)
	if err != nil || maintenancePrincipal != "harpy_cluster_admin" || database != "postgres" {
		return configError(maintenanceKey, "must use harpy_cluster_admin and database postgres")
	}
	targetPrincipal, _, err := postgresDSNPrincipalAndDatabase(target)
	if err != nil || targetPrincipal != maintenancePrincipal {
		return configError(targetKey, "must use the maintenance credential principal")
	}
	return nil
}

func (postgres BootstrapPostgres) EndpointKind() PostgresEndpointKind {
	return postgres.endpointKind
}

func (postgres BootstrapPostgres) MaintenanceDSN() Secret { return postgres.maintenanceDSN }
func (postgres BootstrapPostgres) TargetDSN() Secret      { return postgres.targetDSN }
func (postgres BootstrapPostgres) ConnectTimeout() time.Duration {
	return postgres.connectTimeout
}
func (postgres BootstrapPostgres) CleanupTimeout() time.Duration {
	return postgres.cleanupTimeout
}
