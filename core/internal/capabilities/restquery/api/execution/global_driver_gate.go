package execution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type leasedGlobalDriver struct{ lease *driverLease }

func (driver leasedGlobalDriver) Owner() string {
	if !driver.lease.usable() {
		return ""
	}
	return driver.lease.owner
}

func (driver leasedGlobalDriver) GlobalScope() transaction.GlobalScope {
	if !driver.lease.usable() {
		return transaction.GlobalScope{}
	}
	return driver.lease.global
}

func (driver leasedGlobalDriver) Execute(
	_ context.Context, statement string, arguments ...any,
) error {
	return leaseExecute(driver.lease, statement, arguments...)
}

func (driver leasedGlobalDriver) Query(
	_ context.Context, statement string, arguments ...any,
) (Rows, error) {
	return leaseQuery(driver.lease, statement, arguments...)
}

func (driver leasedGlobalDriver) QueryRow(
	_ context.Context, statement string, arguments ...any,
) Row {
	return leaseQueryRow(driver.lease, statement, arguments...)
}
