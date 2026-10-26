package execution

import "sync/atomic"

type leasedRows struct {
	lease  *driverLease
	rows   Rows
	closed atomic.Bool
}

func (rows *leasedRows) Close() {
	if rows == nil || !rows.closed.CompareAndSwap(false, true) {
		return
	}
	rows.rows.Close()
	rows.lease.release(rows)
}

func (rows *leasedRows) Err() error {
	if rows == nil || !rows.lease.usable() {
		return errExecutionContract
	}
	return rows.rows.Err()
}

func (rows *leasedRows) Next() bool {
	return rows != nil && rows.lease.usable() && rows.rows.Next()
}

func (rows *leasedRows) Scan(destinations ...any) error {
	if rows == nil || !rows.lease.usable() {
		return errExecutionContract
	}
	return rows.rows.Scan(destinations...)
}

type leasedRow struct {
	lease *driverLease
	row   Row
}

func (row leasedRow) Scan(destinations ...any) error {
	if !row.lease.usable() {
		return errExecutionContract
	}
	return row.row.Scan(destinations...)
}

type errorRow struct {
	err error
}

func (row errorRow) Scan(...any) error { return row.err }
