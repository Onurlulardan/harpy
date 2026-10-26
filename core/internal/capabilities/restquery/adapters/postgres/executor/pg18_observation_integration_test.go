//go:build integration

package executor

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgconn"
)

type pg18Call struct {
	kind      string
	statement string
	arguments []any
}

type pg18Observation struct {
	mu             sync.Mutex
	calls          []pg18Call
	databaseErrors []string
	rowsClosedHook func()
	opened         atomic.Int32
	closed         atomic.Int32
}

func (observation *pg18Observation) recordDatabaseError(err error) {
	if err == nil {
		return
	}
	var postgresError *pgconn.PgError
	value := "non-PostgreSQL error: " + err.Error()
	if errors.As(err, &postgresError) {
		value = postgresError.Code + ": " + postgresError.Message
	}
	observation.mu.Lock()
	defer observation.mu.Unlock()
	observation.databaseErrors = append(observation.databaseErrors, value)
}

func (observation *pg18Observation) errorSnapshot() []string {
	observation.mu.Lock()
	defer observation.mu.Unlock()
	return append([]string{}, observation.databaseErrors...)
}

func (observation *pg18Observation) record(
	kind string, statement string, arguments []any,
) {
	observation.mu.Lock()
	defer observation.mu.Unlock()
	observation.calls = append(observation.calls, pg18Call{
		kind: kind, statement: statement,
		arguments: append([]any{}, arguments...),
	})
}

func (observation *pg18Observation) snapshot() []pg18Call {
	observation.mu.Lock()
	defer observation.mu.Unlock()
	result := make([]pg18Call, len(observation.calls))
	copy(result, observation.calls)
	return result
}

func (observation *pg18Observation) setRowsClosedHook(hook func()) {
	observation.mu.Lock()
	defer observation.mu.Unlock()
	observation.rowsClosedHook = hook
}

func (observation *pg18Observation) closeRows() {
	observation.closed.Add(1)
	observation.mu.Lock()
	hook := observation.rowsClosedHook
	observation.mu.Unlock()
	if hook != nil {
		hook()
	}
}
