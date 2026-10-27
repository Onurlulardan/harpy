package config

import "time"

const requiredPostgresCleanupTimeout = 5 * time.Second

func loadPostgresCleanupTimeout(values environmentValues) (time.Duration, error) {
	timeout, err := parseDuration(
		values, "HARPY_POSTGRES_CLEANUP_TIMEOUT", "5s",
		requiredPostgresCleanupTimeout, requiredPostgresCleanupTimeout,
	)
	if err != nil {
		return 0, err
	}
	return timeout, nil
}
