package api

import (
	"context"
	"sync"
)

type notificationsProjectorFake struct {
	mutex *sync.Mutex
	calls []notificationsProjectionCall
	count int
	err   error
}

type notificationsProjectionCall struct {
	actor, tenant string
	limit         int
}

func (fake *notificationsProjectorFake) DrainTenant(_ context.Context,
	actor, tenant string, limit int,
) (int, error) {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	fake.calls = append(fake.calls, notificationsProjectionCall{actor, tenant, limit})
	return fake.count, fake.err
}

type notificationsRealtimeFake struct {
	tenants []string
	drained bool
	forced  bool
	waited  bool
	mutex   *sync.Mutex
}

func (fake *notificationsRealtimeFake) ActiveTenantIDs() []string {
	return append([]string(nil), fake.tenants...)
}

func (fake *notificationsRealtimeFake) BeginDrain() {
	fake.mutex.Lock()
	fake.drained = true
	fake.mutex.Unlock()
}

func (fake *notificationsRealtimeFake) Wait(context.Context) error {
	fake.mutex.Lock()
	fake.waited = true
	fake.mutex.Unlock()
	return nil
}

func (fake *notificationsRealtimeFake) Force() error {
	fake.mutex.Lock()
	fake.forced = true
	fake.mutex.Unlock()
	return nil
}

func projectorFixture(count int, err error,
) (*notificationsProjectorFake, *notificationsRealtimeFake) {
	mutex := new(sync.Mutex)
	return &notificationsProjectorFake{mutex: mutex, count: count, err: err},
		&notificationsRealtimeFake{mutex: mutex, tenants: []string{
			"019d0000-0000-7000-8000-000000000101",
			"019d0000-0000-7000-8000-000000000102",
		}}
}
