package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBindAPIStoreRequiresAPIProfile(t *testing.T) {
	t.Parallel()
	pool := &Pool{profile: ProfileWorker, cleanupTimeout: time.Second}
	called := false
	_, err := BindAPIStore(pool, func(*pgxpool.Pool, time.Duration) (struct{}, error) {
		called = true
		return struct{}{}, nil
	})
	if !errors.Is(err, ErrAPIStoreProfile) || called {
		t.Fatalf("profile fence = %v, factory called = %t", err, called)
	}
}

func TestBindAPIStoreRejectsRawPoolIdentity(t *testing.T) {
	t.Parallel()
	pool := apiStoreTestPool()
	direct, err := BindAPIStore(pool, func(raw *pgxpool.Pool, _ time.Duration) (*pgxpool.Pool, error) {
		return raw, nil
	})
	if direct != nil || !errors.Is(err, ErrAPIStoreBindingInvalid) {
		t.Fatalf("direct raw pool = %v, %v", direct, err)
	}
	boxed, err := BindAPIStore(pool, func(raw *pgxpool.Pool, _ time.Duration) (any, error) {
		return raw, nil
	})
	if boxed != nil || !errors.Is(err, ErrAPIStoreBindingInvalid) {
		t.Fatalf("boxed raw pool = %v, %v", boxed, err)
	}
}

func TestBindAPIStoreBuildsPrivateStore(t *testing.T) {
	t.Parallel()
	type privateStore struct{ cleanup time.Duration }
	pool := apiStoreTestPool()
	store, err := BindAPIStore(pool, func(*pgxpool.Pool, time.Duration) (*privateStore, error) {
		return &privateStore{cleanup: 3 * time.Second}, nil
	})
	if err != nil || store == nil || store.cleanup != 3*time.Second {
		t.Fatalf("private store = %#v, %v", store, err)
	}
	want := errors.New("factory failed")
	_, err = BindAPIStore(pool, func(*pgxpool.Pool, time.Duration) (*privateStore, error) {
		return nil, want
	})
	if !errors.Is(err, want) {
		t.Fatalf("factory error = %v", err)
	}
}

func TestBindAPIStoreRejectsClosedPoolAndNilFactory(t *testing.T) {
	t.Parallel()
	pool := apiStoreTestPool()
	if _, err := BindAPIStore[any](pool, nil); !errors.Is(err, ErrAPIStoreBindingInvalid) {
		t.Fatalf("nil factory error = %v", err)
	}
	pool.closed.Store(true)
	if _, err := BindAPIStore[any](pool, nil); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("closed pool error = %v", err)
	}
}

func apiStoreTestPool() *Pool {
	return &Pool{
		profile: ProfileAPI, cleanupTimeout: 3 * time.Second,
		contract: profileContract{runtime: true},
	}
}
