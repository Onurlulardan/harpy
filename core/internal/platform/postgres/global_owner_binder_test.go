package postgres

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestGlobalOwnerBinderIsBillingOnlyAndTenantDistinct(t *testing.T) {
	type target struct{ database BoundExecutor }
	build := func(executor BoundExecutor) target {
		return target{database: executor}
	}
	if _, err := NewGlobalOwnerBinder("forms", build); !errors.Is(
		err, ErrOwnerBindingInvalid,
	) {
		t.Fatalf("non-billing global owner error = %v", err)
	}
	global, err := NewGlobalOwnerBinder("billing", build)
	if err != nil || global.Owner() != "billing" {
		t.Fatalf("global billing binder = %#v, %v", global, err)
	}
	if reflect.TypeOf(global) == reflect.TypeOf(OwnerBinder[target]{}) {
		t.Fatal("global and tenant owner binders share an identity")
	}
}

func TestGlobalOwnerBinderBindsOnlyPrivateRepository(t *testing.T) {
	transaction := &RuntimeTransaction{
		bindings: newBindingState(), cleanup: time.Second,
	}
	transaction.open.Store(true)
	type repository struct{ database BoundExecutor }
	binder, err := NewGlobalOwnerBinder(
		"billing", func(executor BoundExecutor) *repository {
			return &repository{database: executor}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := BindGlobalOwner(transaction, binder)
	if err != nil || result == nil || !result.database.usable() {
		t.Fatalf("global private repository = %#v, %v", result, err)
	}
	raw, _ := NewGlobalOwnerBinder(
		"billing", func(executor BoundExecutor) BoundExecutor { return executor },
	)
	if _, err := BindGlobalOwner(transaction, raw); !errors.Is(
		err, ErrOwnerBindingInvalid,
	) {
		t.Fatalf("raw global executor identity error = %v", err)
	}
}
