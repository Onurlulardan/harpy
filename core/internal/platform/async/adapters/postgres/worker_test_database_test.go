package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var errWorkerTestDatabase = errors.New("worker test database was called")

type workerTestDatabase struct {
	transaction *workerTestTransaction
	beginCalls  int
	queryRows   pgx.Rows
	querySQL    string
	queryArgs   []any
}

func (database *workerTestDatabase) Query(
	_ context.Context,
	statement string,
	arguments ...any,
) (pgx.Rows, error) {
	database.querySQL = statement
	database.queryArgs = append([]any(nil), arguments...)
	if database.queryRows != nil {
		return database.queryRows, nil
	}
	return nil, errWorkerTestDatabase
}

func (database *workerTestDatabase) BeginTx(
	_ context.Context,
	_ pgx.TxOptions,
) (pgx.Tx, error) {
	database.beginCalls++
	if database.transaction == nil {
		return nil, errWorkerTestDatabase
	}
	return database.transaction, nil
}

type workerTestTransaction struct {
	queryRows     pgx.Rows
	queryErr      error
	commitErr     error
	rollbackCalls int
	calls         []string
}

func (transaction *workerTestTransaction) Begin(context.Context) (pgx.Tx, error) {
	return nil, errWorkerTestDatabase
}
func (transaction *workerTestTransaction) Commit(context.Context) error {
	transaction.calls = append(transaction.calls, "commit")
	return transaction.commitErr
}
func (transaction *workerTestTransaction) Rollback(context.Context) error {
	transaction.rollbackCalls++
	transaction.calls = append(transaction.calls, "rollback")
	return nil
}
func (transaction *workerTestTransaction) CopyFrom(
	context.Context,
	pgx.Identifier,
	[]string,
	pgx.CopyFromSource,
) (int64, error) {
	return 0, errWorkerTestDatabase
}
func (transaction *workerTestTransaction) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}
func (transaction *workerTestTransaction) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}
func (transaction *workerTestTransaction) Prepare(
	context.Context,
	string,
	string,
) (*pgconn.StatementDescription, error) {
	return nil, errWorkerTestDatabase
}
func (transaction *workerTestTransaction) Exec(
	_ context.Context,
	statement string,
	_ ...any,
) (pgconn.CommandTag, error) {
	if statement == assertJobAuthoritySQL {
		transaction.calls = append(transaction.calls, "assert authority")
	} else {
		transaction.calls = append(transaction.calls, "unexpected execute")
	}
	return pgconn.CommandTag{}, nil
}
func (transaction *workerTestTransaction) Query(
	_ context.Context,
	statement string,
	_ ...any,
) (pgx.Rows, error) {
	if statement == claimJobsSQL {
		transaction.calls = append(transaction.calls, "lock jobs")
	} else {
		transaction.calls = append(transaction.calls, "unexpected query")
	}
	return transaction.queryRows, transaction.queryErr
}
func (transaction *workerTestTransaction) QueryRow(
	context.Context,
	string,
	...any,
) pgx.Row {
	return workerTestRow{}
}
func (transaction *workerTestTransaction) Conn() *pgx.Conn { return nil }

type workerTestRow struct{}

func (workerTestRow) Scan(...any) error { return errWorkerTestDatabase }

type workerTestRows struct {
	remaining  int
	scanErr    error
	iterateErr error
	closed     bool
	values     [][]any
	index      int
}

func (rows *workerTestRows) Close() { rows.closed = true }
func (rows *workerTestRows) Err() error {
	if rows.closed {
		return rows.iterateErr
	}
	return nil
}
func (rows *workerTestRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (rows *workerTestRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (rows *workerTestRows) Next() bool {
	if len(rows.values) > 0 {
		if rows.closed || rows.index >= len(rows.values) {
			rows.closed = true
			return false
		}
		rows.index++
		return true
	}
	if rows.closed || rows.remaining == 0 {
		rows.closed = true
		return false
	}
	rows.remaining--
	return true
}
func (rows *workerTestRows) Scan(destinations ...any) error {
	if rows.scanErr != nil {
		rows.closed = true
		return rows.scanErr
	}
	if len(rows.values) == 0 || rows.index == 0 ||
		len(destinations) != len(rows.values[rows.index-1]) {
		return errWorkerTestDatabase
	}
	for index, value := range rows.values[rows.index-1] {
		if err := assignWorkerTestValue(destinations[index], value); err != nil {
			return err
		}
	}
	return nil
}
func (rows *workerTestRows) Values() ([]any, error) { return nil, errWorkerTestDatabase }
func (rows *workerTestRows) RawValues() [][]byte    { return nil }
func (rows *workerTestRows) Conn() *pgx.Conn        { return nil }
