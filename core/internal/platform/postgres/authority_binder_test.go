package postgres

import (
	"errors"
	"testing"
	"time"
)

func TestAuthorityBinderRejectsRawExecutorIdentity(t *testing.T) {
	transaction := &RuntimeTransaction{
		bindings: newBindingState(), cleanup: time.Second,
	}
	transaction.open.Store(true)
	direct, err := NewAuthorityBinder("platform_async", func(value BoundExecutor) BoundExecutor {
		return value
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindAuthority(transaction, direct); !errors.Is(err, ErrAuthorityBindingInvalid) {
		t.Fatalf("direct executor identity error = %v", err)
	}
	boxed, err := NewAuthorityBinder("platform_async", func(value BoundExecutor) any {
		return value
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindAuthority(transaction, boxed); !errors.Is(err, ErrAuthorityBindingInvalid) {
		t.Fatalf("boxed executor identity error = %v", err)
	}
}

func TestAuthorityBinderAllowsPrivateRepository(t *testing.T) {
	type repository struct{ database BoundExecutor }
	transaction := &RuntimeTransaction{
		bindings: newBindingState(), cleanup: time.Second,
	}
	transaction.open.Store(true)
	binder, err := NewAuthorityBinder("platform_async", func(value BoundExecutor) *repository {
		return &repository{database: value}
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := BindAuthority(transaction, binder)
	if err != nil || result == nil || !result.database.usable() {
		t.Fatalf("private repository binding = %#v, %v", result, err)
	}
}

func TestAuthorityBinderValidatesOwnerAndBuilder(t *testing.T) {
	t.Parallel()
	if _, err := NewAuthorityBinder[any]("invalid-owner", func(BoundExecutor) any {
		return struct{}{}
	}); err != ErrAuthorityBindingInvalid {
		t.Fatalf("invalid owner error = %v", err)
	}
	if _, err := NewAuthorityBinder[any]("platform_async", nil); err != ErrAuthorityBindingInvalid {
		t.Fatalf("nil builder error = %v", err)
	}
}
