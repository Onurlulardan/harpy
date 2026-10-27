package postgres

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

const documentsRecoveryOwner = "documents"

type DocumentsRecoveryDatabase interface {
	Execute(context.Context, string, ...any) (CommandResult, error)
	Query(context.Context, string, ...any) (Rows, error)
	QueryRow(context.Context, string, ...any) Row
}

type DocumentsRecoveryTransactions interface {
	WithinTenant(
		context.Context,
		TransactionMode,
		uuid.UUID,
		uuid.UUID,
		[32]byte,
		func(DocumentsRecoveryDatabase) error,
	) error
	WithinTenantSnapshot(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		[32]byte,
		string,
		func(DocumentsRecoveryDatabase) error,
	) error
}

type DocumentsRecoverySnapshotImporter func(
	context.Context, DocumentsRecoveryDatabase, string,
) error

type documentsRecoveryTransactions struct {
	pool             *Pool
	snapshotImporter DocumentsRecoverySnapshotImporter
}
type documentsRecoveryDatabase struct{ executor BoundExecutor }

func (database documentsRecoveryDatabase) Execute(
	ctx context.Context, statement string, arguments ...any,
) (CommandResult, error) {
	return database.executor.Execute(ctx, statement, arguments...)
}
func (database documentsRecoveryDatabase) Query(
	ctx context.Context, statement string, arguments ...any,
) (Rows, error) {
	return database.executor.Query(ctx, statement, arguments...)
}
func (database documentsRecoveryDatabase) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) Row {
	return database.executor.QueryRow(ctx, statement, arguments...)
}

func (runner documentsRecoveryTransactions) WithinTenant(
	ctx context.Context,
	mode TransactionMode,
	environmentID uuid.UUID,
	tenantID uuid.UUID,
	tenantSetSHA256 [32]byte,
	callback func(DocumentsRecoveryDatabase) error,
) error {
	if callback == nil || environmentID == uuid.Nil || tenantID == uuid.Nil {
		return ErrOwnerBindingInvalid
	}
	transaction, err := runner.pool.beginProfileTransaction(ctx, mode)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	purpose := "backup"
	if runner.pool.profile == ProfileDatabaseRestore {
		purpose = "restore"
	}
	record, digest, err := transaction.resolveRecoveryTenant(ctx, purpose)
	if err != nil || record.environmentID != environmentID || record.tenantID != tenantID ||
		digest != tenantSetSHA256 {
		return ErrTenantIdentityMismatch
	}
	if err = transaction.BindTenantSearchPath(ctx, record.tenant); err != nil {
		return err
	}
	binder, err := NewOwnerBinder(documentsRecoveryOwner,
		func(executor BoundExecutor) documentsRecoveryDatabase {
			return documentsRecoveryDatabase{executor: executor}
		})
	if err != nil {
		return err
	}
	database, err := BindOwner(transaction, binder)
	if err != nil {
		return err
	}
	if err = callback(database); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

func (runner documentsRecoveryTransactions) WithinTenantSnapshot(
	ctx context.Context,
	environmentID uuid.UUID,
	tenantID uuid.UUID,
	tenantSetSHA256 [32]byte,
	snapshotID string,
	callback func(DocumentsRecoveryDatabase) error,
) error {
	if callback == nil || environmentID == uuid.Nil || tenantID == uuid.Nil ||
		!recoverySnapshotPattern.MatchString(snapshotID) ||
		runner.pool.profile != ProfileDatabaseBackupExecutor {
		return ErrOwnerBindingInvalid
	}
	transaction, err := runner.pool.beginProfileTransaction(
		ctx, TransactionSnapshotQuery,
	)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	binder, err := NewOwnerBinder(
		documentsRecoveryOwner,
		func(executor BoundExecutor) documentsRecoveryDatabase {
			return documentsRecoveryDatabase{executor: executor}
		},
	)
	if err != nil {
		return err
	}
	database, err := BindOwner(transaction, binder)
	if err != nil {
		return err
	}
	if err = runner.snapshotImporter(ctx, database, snapshotID); err != nil {
		return err
	}
	record, digest, err := transaction.resolveRecoveryTenant(ctx, "backup")
	if err != nil || record.environmentID != environmentID ||
		record.tenantID != tenantID || digest != tenantSetSHA256 {
		return ErrTenantIdentityMismatch
	}
	if err = transaction.BindTenantSearchPath(ctx, record.tenant); err != nil {
		return err
	}
	if err = callback(database); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

func BindDocumentsRecoveryStore[T any](
	pool *Pool,
	snapshotImporter DocumentsRecoverySnapshotImporter,
	factory func(DocumentsRecoveryTransactions) (T, error),
) (T, error) {
	var zero T
	if !recoveryProfile(pool) || snapshotImporter == nil || factory == nil {
		return zero, fmt.Errorf("documents recovery store binding is unavailable")
	}
	runner := documentsRecoveryTransactions{
		pool: pool, snapshotImporter: snapshotImporter,
	}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
