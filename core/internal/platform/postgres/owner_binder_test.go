package postgres

import (
	"errors"
	"testing"
	"time"
)

func TestOwnerBinderRejectsRawExecutorIdentity(t *testing.T) {
	transaction := &RuntimeTransaction{
		bindings: newBindingState(), cleanup: time.Second,
	}
	transaction.open.Store(true)
	direct, err := NewOwnerBinder("iam", func(executor BoundExecutor) BoundExecutor {
		return executor
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindOwner(transaction, direct); !errors.Is(err, ErrOwnerBindingInvalid) {
		t.Fatalf("direct executor identity error = %v", err)
	}
	boxed, err := NewOwnerBinder("iam", func(executor BoundExecutor) any {
		return executor
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindOwner(transaction, boxed); !errors.Is(err, ErrOwnerBindingInvalid) {
		t.Fatalf("boxed executor identity error = %v", err)
	}
}

func TestOwnerBinderAllowsPrivateRepository(t *testing.T) {
	type repository struct{ database BoundExecutor }
	transaction := &RuntimeTransaction{
		bindings: newBindingState(), cleanup: time.Second,
	}
	transaction.open.Store(true)
	binder, err := NewOwnerBinder("iam", func(executor BoundExecutor) *repository {
		return &repository{database: executor}
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := BindOwner(transaction, binder)
	if err != nil || result == nil || !result.database.usable() {
		t.Fatalf("private repository binding = %#v, %v", result, err)
	}
}
