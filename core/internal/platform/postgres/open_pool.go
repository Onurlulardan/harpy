package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(
	ctx context.Context,
	settings config.Postgres,
	profile Profile,
	embeddedContract schemacontract.EmbeddedContract,
) (*Pool, error) {
	contract, err := profile.contract()
	if err != nil {
		return nil, err
	}
	if !embeddedContract.IsVerified() || embeddedContract.Value().Validate() != nil {
		return nil, fmt.Errorf("postgres binary schema contract is not verified")
	}
	binaryContract := embeddedContract.Value()
	poolConfig, err := pgxpool.ParseConfig(settings.DSN().Value())
	if err != nil {
		return nil, fmt.Errorf("postgres pool config is invalid")
	}
	configurePool(poolConfig, settings, contract)
	probeConfig := poolConfig.ConnConfig.Copy()
	delete(probeConfig.RuntimeParams, "role")
	collector := &startupCollector{}
	installPoolHooks(poolConfig, settings, profile, contract, binaryContract, collector)
	inner, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres pool creation failed")
	}
	pool := &Pool{
		inner: inner, profile: profile, contract: contract,
		cleanupTimeout: settings.CleanupTimeout(), binaryContract: binaryContract,
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
	pool.startup.binaryContractDigest = embeddedContract.DigestSHA256()
	return pool, nil
}
