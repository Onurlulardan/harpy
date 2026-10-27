package postgres

import (
	"context"
	"fmt"
	"regexp"
)

var recoverySnapshotPattern = regexp.MustCompile(`^[0-9]+-[0-9A-F]+-[0-9]+$`)

type RecoverySnapshot struct {
	transaction *RuntimeTransaction
	identifier  string
}

type RecoverySnapshotDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type RecoverySnapshotExporter func(
	context.Context, RecoverySnapshotDatabase,
) (string, error)

func (pool *Pool) BeginRecoverySnapshot(
	ctx context.Context,
	exporter RecoverySnapshotExporter,
) (*RecoverySnapshot, error) {
	if pool == nil || pool.profile != ProfileDatabaseBackupExecutor ||
		exporter == nil {
		return nil, ErrRuntimeProfile
	}
	transaction, err := pool.beginProfileTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return nil, err
	}
	binder, err := NewOwnerBinder(
		"platformops",
		func(executor BoundExecutor) RecoverySnapshotDatabase {
			return executor
		},
	)
	if err != nil {
		_ = transaction.Rollback()
		return nil, err
	}
	database, err := BindOwner(transaction, binder)
	if err != nil {
		_ = transaction.Rollback()
		return nil, err
	}
	identifier, err := exporter(ctx, database)
	if err != nil || !recoverySnapshotPattern.MatchString(identifier) {
		_ = transaction.Rollback()
		return nil, fmt.Errorf("recovery snapshot export failed")
	}
	return &RecoverySnapshot{
		transaction: transaction,
		identifier:  identifier,
	}, nil
}

func (snapshot *RecoverySnapshot) Identifier() string {
	if snapshot == nil {
		return ""
	}
	return snapshot.identifier
}

func (snapshot *RecoverySnapshot) Close() error {
	if snapshot == nil || snapshot.transaction == nil {
		return ErrTransactionClosed
	}
	return snapshot.transaction.Rollback()
}
