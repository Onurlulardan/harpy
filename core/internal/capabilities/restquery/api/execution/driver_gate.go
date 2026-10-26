package execution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type leasedDriver struct {
	lease *driverLease
}

func (driver leasedDriver) Owner() string {
	if !driver.lease.usable() {
		return ""
	}
	return driver.lease.owner
}

func (driver leasedDriver) Scope() transaction.TenantScope {
	if !driver.lease.usable() {
		return transaction.TenantScope{}
	}
	return driver.lease.scope
}

func (driver leasedDriver) Execute(
	_ context.Context,
	statement string,
	arguments ...any,
) error {
	return leaseExecute(driver.lease, statement, arguments...)
}

func (driver leasedDriver) Query(
	_ context.Context,
	statement string,
	arguments ...any,
) (Rows, error) {
	return leaseQuery(driver.lease, statement, arguments...)
}

func (driver leasedDriver) QueryRow(
	_ context.Context,
	statement string,
	arguments ...any,
) Row {
	return leaseQueryRow(driver.lease, statement, arguments...)
}
