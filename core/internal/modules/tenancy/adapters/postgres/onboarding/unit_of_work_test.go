package onboarding

import (
	"context"
	"errors"
	"testing"

	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func TestActivationUnitOfWorkCommitsAndClosesExactTransaction(t *testing.T) {
	transaction := unitTransactionFixture()
	commits, rollbacks := 0, 0
	unit := &activationUnitOfWork{begin: func(context.Context) (activationBoundary, error) {
		return activationBoundary{transaction: transaction,
			commit:   func(context.Context) error { commits++; return nil },
			rollback: func() error { rollbacks++; return nil }}, nil
	}}
	var captured onboardingapp.ActivationTransaction
	err := unit.Run(context.Background(), func(
		_ context.Context,
		bound onboardingapp.ActivationTransaction,
	) error {
		captured = bound
		if bound != transaction {
			return errors.New("transaction identity changed")
		}
		return nil
	})
	if err != nil || commits != 1 || rollbacks != 0 || captured != transaction ||
		transaction.usable() {
		t.Fatalf("commit=%d rollback=%d usable=%t err=%v",
			commits, rollbacks, transaction.usable(), err)
	}
}

func TestActivationUnitOfWorkRollsBackAndClosesOnCallbackFailure(t *testing.T) {
	transaction := unitTransactionFixture()
	commits, rollbacks := 0, 0
	unit := &activationUnitOfWork{begin: func(context.Context) (activationBoundary, error) {
		return activationBoundary{transaction: transaction,
			commit:   func(context.Context) error { commits++; return nil },
			rollback: func() error { rollbacks++; return nil }}, nil
	}}
	crash := errors.New("owner callback crashed")
	err := unit.Run(context.Background(), func(
		_ context.Context,
		bound onboardingapp.ActivationTransaction,
	) error {
		if bound != transaction {
			return errors.New("transaction identity changed")
		}
		return crash
	})
	if !errors.Is(err, crash) || commits != 0 || rollbacks != 1 || transaction.usable() {
		t.Fatalf("commit=%d rollback=%d usable=%t err=%v",
			commits, rollbacks, transaction.usable(), err)
	}
}

func TestActivationUnitOfWorkRollsBackOnPanic(t *testing.T) {
	transaction := unitTransactionFixture()
	rollbacks := 0
	unit := &activationUnitOfWork{begin: func(context.Context) (activationBoundary, error) {
		return activationBoundary{transaction: transaction,
			commit:   func(context.Context) error { return nil },
			rollback: func() error { rollbacks++; return nil }}, nil
	}}
	func() {
		defer func() { _ = recover() }()
		_ = unit.Run(context.Background(), func(
			context.Context,
			onboardingapp.ActivationTransaction,
		) error {
			panic("crash")
		})
	}()
	if rollbacks != 1 || transaction.usable() {
		t.Fatalf("panic rollback=%d usable=%t", rollbacks, transaction.usable())
	}
}

func unitTransactionFixture() *activationTransaction {
	result := &activationTransaction{store: &postgresStore{
		queryRow: func(context.Context, string, ...any) Row { return nil },
		execute:  noopExecute,
	}, bind: func(onboardingapp.ActivationOwner) (onboardingapp.ActivationOwnerExecutor, error) {
		return onboardingapp.ActivationOwnerExecutor{}, nil
	}, bindSchema: func(context.Context, string) error { return nil }}
	result.open.Store(true)
	return result
}
