package execution

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type driverLease struct {
	context context.Context
	driver  QueryDriver
	owner   string
	scope   transaction.TenantScope
	global  transaction.GlobalScope
	active  atomic.Bool
	mutex   sync.Mutex
	rows    map[*leasedRows]struct{}
}

func newGlobalDriverLease(
	ctx context.Context,
	driver GlobalSnapshotDriver,
	owner string,
	scope transaction.GlobalScope,
) (*driverLease, error) {
	if ctx == nil || ctx.Err() != nil || nilInterface(driver) ||
		driver.Owner() != owner || driver.GlobalScope() != scope {
		return nil, errExecutionContract
	}
	lease := &driverLease{
		context: ctx, driver: driver, owner: owner, global: scope,
		rows: make(map[*leasedRows]struct{}),
	}
	lease.active.Store(true)
	context.AfterFunc(ctx, lease.close)
	return lease, nil
}

func newDriverLease(
	ctx context.Context,
	driver SnapshotDriver,
	owner string,
	scope transaction.TenantScope,
) (*driverLease, error) {
	if ctx == nil || ctx.Err() != nil || nilInterface(driver) ||
		driver.Owner() != owner || driver.Scope() != scope {
		return nil, errExecutionContract
	}
	lease := &driverLease{
		context: ctx, driver: driver, owner: owner, scope: scope,
		rows: make(map[*leasedRows]struct{}),
	}
	lease.active.Store(true)
	context.AfterFunc(ctx, lease.close)
	return lease, nil
}

func (lease *driverLease) usable() bool {
	return lease != nil && lease.active.Load() && lease.context != nil &&
		lease.context.Err() == nil
}

func (lease *driverLease) track(rows *leasedRows) bool {
	lease.mutex.Lock()
	defer lease.mutex.Unlock()
	if !lease.active.Load() {
		return false
	}
	lease.rows[rows] = struct{}{}
	return true
}

func (lease *driverLease) release(rows *leasedRows) {
	lease.mutex.Lock()
	delete(lease.rows, rows)
	lease.mutex.Unlock()
}

func (lease *driverLease) close() {
	if lease == nil || !lease.active.Swap(false) {
		return
	}
	lease.mutex.Lock()
	rows := make([]*leasedRows, 0, len(lease.rows))
	for current := range lease.rows {
		rows = append(rows, current)
	}
	lease.mutex.Unlock()
	for _, current := range rows {
		current.Close()
	}
}
