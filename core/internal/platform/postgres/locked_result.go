package postgres

import (
	"sync"
)

type lockedRows struct {
	mutex     sync.Mutex
	rows      Rows
	operation *boundOperation
	closed    bool
}

func newLockedRows(rows Rows, operation *boundOperation) *lockedRows {
	result := &lockedRows{rows: rows, operation: operation}
	operation.installAbort(result.abort)
	return result
}

func (rows *lockedRows) Close() { rows.close() }

func (rows *lockedRows) Err() error {
	rows.mutex.Lock()
	err := rows.rows.Err()
	closed := rows.closed
	rows.mutex.Unlock()
	if err != nil || closed {
		rows.operation.finish()
	}
	return err
}

func (rows *lockedRows) Next() bool {
	rows.mutex.Lock()
	if rows.closed {
		rows.mutex.Unlock()
		return false
	}
	next := rows.rows.Next()
	if !next {
		rows.closed = true
	}
	rows.mutex.Unlock()
	if !next {
		rows.operation.finish()
	}
	return next
}

func (rows *lockedRows) Scan(destinations ...any) error {
	rows.mutex.Lock()
	if rows.closed {
		rows.mutex.Unlock()
		return ErrTransactionClosed
	}
	err := rows.rows.Scan(destinations...)
	rows.mutex.Unlock()
	if err != nil {
		rows.close()
	}
	return err
}

func (rows *lockedRows) abort() { rows.close() }

func (rows *lockedRows) close() {
	rows.mutex.Lock()
	if !rows.closed {
		rows.rows.Close()
		rows.closed = true
	}
	rows.mutex.Unlock()
	rows.operation.finish()
}
