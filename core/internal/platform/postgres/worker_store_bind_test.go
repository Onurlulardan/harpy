package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBindWorkerStoreRequiresWorkerProfile(t *testing.T) {
	t.Parallel()
	pool := &Pool{profile: ProfileAPI, cleanupTimeout: time.Second}
	called := false
	_, err := BindWorkerStore(pool, func(*pgxpool.Pool, time.Duration) (struct{}, error) {
		called = true
		return struct{}{}, nil
	})
	if !errors.Is(err, ErrWorkerStoreProfile) || called {
		t.Fatalf("profile fence = %v, factory called = %t", err, called)
	}
}

func TestBindWorkerStoreRejectsRawPoolIdentity(t *testing.T) {
	t.Parallel()
	pool := workerStoreTestPool()
	direct, err := BindWorkerStore(pool, func(raw *pgxpool.Pool, _ time.Duration) (*pgxpool.Pool, error) {
		return raw, nil
	})
	if direct != nil || !errors.Is(err, ErrWorkerStoreBindingInvalid) {
		t.Fatalf("direct raw pool = %v, %v", direct, err)
	}
}

func TestBindWorkerStoreBuildsPrivateStore(t *testing.T) {
	t.Parallel()
	type privateStore struct{ cleanup time.Duration }
	pool := workerStoreTestPool()
	store, err := BindWorkerStore(pool, func(*pgxpool.Pool, time.Duration) (*privateStore, error) {
		return &privateStore{cleanup: 3 * time.Second}, nil
	})
	if err != nil || store == nil || store.cleanup != 3*time.Second {
		t.Fatalf("private store = %#v, %v", store, err)
	}
}

func workerStoreTestPool() *Pool {
	return &Pool{
		profile: ProfileWorker, cleanupTimeout: 3 * time.Second,
		contract: profileContract{runtime: true},
	}
}
