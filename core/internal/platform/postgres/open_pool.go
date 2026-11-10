package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(
	ctx context.Context,
	settings config.Postgres,
	profile Profile,
) (*Pool, error) {
	contract, err := profile.contract()
	if err != nil {
		return nil, err
	}
	binaryContract, err := loadEmbeddedBinaryContract()
	if err != nil {
		return nil, err
	}
	poolConfig, err := pgxpool.ParseConfig(settings.DSN().Value())
	if err != nil {
		return nil, fmt.Errorf("postgres pool config is invalid")
	}
	configurePool(poolConfig, settings, contract)
	probeConfig := poolConfig.ConnConfig.Copy()
	delete(probeConfig.RuntimeParams, "role")
	collector := &startupCollector{}
	installPoolHooks(poolConfig, settings, profile, contract, collector)
	inner, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres pool creation failed")
	}
	pool := &Pool{
		inner: inner, profile: profile, contract: contract,
		binaryContract: binaryContract, cleanupTimeout: settings.CleanupTimeout(),
		probeConfig: probeConfig,
	}
	connection, err := inner.Acquire(ctx)
	if err != nil {
		inner.Close()
		return nil, fmt.Errorf("%w: connection unavailable: %w", ErrStartupPreflight, err)
	}
	connection.Release()
	result, err := collector.value()
	if err != nil {
		inner.Close()
		return nil, err
	}
	pool.startup = result
	return pool, nil
}
