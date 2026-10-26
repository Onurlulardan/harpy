package postgres

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	inner          *pgxpool.Pool
	profile        Profile
	contract       profileContract
	startup        StartupResult
	cleanupTimeout time.Duration
	binaryContract runtimev1.BinarySchemaContract
	probeConfig    *pgx.ConnConfig
	closed         atomic.Bool
	forceOnce      sync.Once
}

// RuntimeOwnerDatabase is an opaque capability marker used by module roots.
// Concrete owner binding remains inside PostgreSQL adapters.
func (*Pool) RuntimeOwnerDatabase() {}

func (pool *Pool) StartupResult() StartupResult {
	if pool == nil {
		return StartupResult{}
	}
	return pool.startup
}

func (pool *Pool) Ping(ctx context.Context) error {
	if pool == nil || pool.closed.Load() {
		return ErrPoolClosed
	}
	return pool.inner.Ping(ctx)
}

func (pool *Pool) Close() {
	if pool == nil || !pool.closed.CompareAndSwap(false, true) {
		return
	}
	pool.inner.Close()
}

// ForceClose rejects new wrapper operations immediately and starts pgxpool
// shutdown without waiting for checked-out connections to return.
func (pool *Pool) ForceClose() {
	if pool == nil {
		return
	}
	alreadyClosing := pool.closed.Swap(true)
	pool.forceOnce.Do(func() {
		pool.inner.Reset()
		if !alreadyClosing {
			go pool.inner.Close()
		}
	})
}

func (pool *Pool) runtimeCapable() bool {
	return pool != nil && !pool.closed.Load() && pool.contract.runtime
}

func (pool *Pool) profileTransactionCapable() bool {
	return pool != nil && !pool.closed.Load() && pool.contract.profileTx
}
