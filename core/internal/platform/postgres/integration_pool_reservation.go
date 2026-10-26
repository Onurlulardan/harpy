//go:build integration

package postgres

import "context"

func ReserveIntegrationConnection(
	ctx context.Context,
	pool *Pool,
) (func(), error) {
	if pool == nil || pool.closed.Load() {
		return nil, ErrPoolClosed
	}
	connection, err := pool.inner.Acquire(ctx)
	if err != nil {
		return nil, ErrPoolClosed
	}
	return connection.Release, nil
}
