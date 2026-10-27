package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBindAsyncWorkerStoreRequiresWorkerProfile(t *testing.T) {
	t.Parallel()
	pool := &Pool{profile: ProfileAPI, cleanupTimeout: time.Second}
	called := false
	_, err := BindAsyncWorkerStore(pool, func(*pgxpool.Pool, time.Duration) struct{} {
		called = true
		return struct{}{}
	})
	if !errors.Is(err, ErrAsyncWorkerProfile) || called {
		t.Fatalf("profile fence = %v, factory called = %t", err, called)
	}
}

func TestBindAsyncWorkerStoreRejectsRawPoolIdentity(t *testing.T) {
	t.Parallel()
	pool := &Pool{profile: ProfileWorker, cleanupTimeout: time.Second}
	direct, err := BindAsyncWorkerStore(
		pool,
		func(raw *pgxpool.Pool, _ time.Duration) *pgxpool.Pool { return raw },
	)
	if direct != nil || !errors.Is(err, ErrAsyncWorkerBindingInvalid) {
		t.Fatalf("direct raw pool = %v, %v", direct, err)
	}
	boxed, err := BindAsyncWorkerStore(
		pool,
		func(raw *pgxpool.Pool, _ time.Duration) any { return raw },
	)
	if boxed != nil || !errors.Is(err, ErrAsyncWorkerBindingInvalid) {
		t.Fatalf("boxed raw pool = %v, %v", boxed, err)
	}
}

func TestBindAsyncWorkerStoreBuildsPrivateStore(t *testing.T) {
	t.Parallel()
	type privateStore struct {
		database *pgxpool.Pool
		cleanup  time.Duration
	}
	pool := &Pool{profile: ProfileWorker, cleanupTimeout: 3 * time.Second}
	store, err := BindAsyncWorkerStore(
		pool,
		func(raw *pgxpool.Pool, cleanup time.Duration) *privateStore {
			return &privateStore{database: raw, cleanup: cleanup}
		},
	)
	if err != nil || store == nil || store.cleanup != 3*time.Second {
		t.Fatalf("private store = %#v, %v", store, err)
	}
}

func TestBindAsyncWorkerStoreRejectsClosedPoolAndNilFactory(t *testing.T) {
	t.Parallel()
	pool := &Pool{profile: ProfileWorker, cleanupTimeout: time.Second}
	if _, err := BindAsyncWorkerStore[any](pool, nil); !errors.Is(
		err, ErrAsyncWorkerBindingInvalid,
	) {
		t.Fatalf("nil factory error = %v", err)
	}
	pool.closed.Store(true)
	if _, err := BindAsyncWorkerStore[any](pool, nil); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("closed pool error = %v", err)
	}
}
