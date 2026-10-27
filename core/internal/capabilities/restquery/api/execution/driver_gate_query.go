package execution

func leaseExecute(lease *driverLease, statement string, arguments ...any) error {
	if !lease.usable() || statement == "" {
		return errExecutionContract
	}
	return lease.driver.Execute(lease.context, statement, arguments...)
}

func leaseQuery(lease *driverLease, statement string, arguments ...any) (Rows, error) {
	if !lease.usable() || statement == "" {
		return nil, errExecutionContract
	}
	rows, err := lease.driver.Query(lease.context, statement, arguments...)
	if err != nil {
		return nil, err
	}
	if nilInterface(rows) {
		return nil, errExecutionContract
	}
	leased := &leasedRows{lease: lease, rows: rows}
	if !lease.track(leased) {
		rows.Close()
		return nil, errExecutionContract
	}
	return leased, nil
}

func leaseQueryRow(lease *driverLease, statement string, arguments ...any) Row {
	if !lease.usable() || statement == "" {
		return errorRow{err: errExecutionContract}
	}
	row := lease.driver.QueryRow(lease.context, statement, arguments...)
	if nilInterface(row) {
		return errorRow{err: errExecutionContract}
	}
	return leasedRow{lease: lease, row: row}
}
