package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

type storeFixture struct {
	contactSubmissions int
	bug                domain.BugReport
	lastTenant         domain.ID
}

func (store *storeFixture) ContactSubmit(context.Context, ContactSubmission) (domain.ContactMessage, bool, error) {
	store.contactSubmissions++
	return domain.ContactMessage{}, false, nil
}
func (*storeFixture) ContactsList(context.Context, Access, domain.ContactFilter) (domain.ContactPage, error) {
	return domain.ContactPage{}, nil
}
func (*storeFixture) ContactStatsRead(context.Context, Access, time.Time) (domain.ContactStats, error) {
	return domain.ContactStats{}, nil
}
func (*storeFixture) ContactGet(context.Context, Access, domain.ID) (domain.ContactMessage, error) {
	return domain.ContactMessage{}, nil
}
func (*storeFixture) ContactUpdate(context.Context, Access, domain.ContactUpdate) (domain.ContactMessage, bool, error) {
	return domain.ContactMessage{}, false, nil
}
func (store *storeFixture) BugCreate(context.Context, BugSubmission) (domain.BugReport, bool, error) {
	return store.bug, false, nil
}
func (*storeFixture) BugsList(context.Context, Access, domain.BugFilter) (domain.BugPage, error) {
	return domain.BugPage{}, nil
}
func (*storeFixture) BugStatsRead(context.Context, Access, domain.ID) (domain.BugStats, error) {
	return domain.BugStats{}, nil
}
func (store *storeFixture) BugGet(_ context.Context, _ Access, _ domain.ID,
	tenantID domain.ID,
) (domain.BugReport, error) {
	store.lastTenant = tenantID
	return store.bug, nil
}
func (*storeFixture) BugUpdate(context.Context, Access, domain.BugUpdate) (domain.BugReport, bool, error) {
	return domain.BugReport{}, false, nil
}

type clockFixture struct{ value time.Time }

func (clock clockFixture) Now() time.Time { return clock.value }

type idFixture struct{ value domain.ID }

func (ids idFixture) NewV7() (domain.ID, error) { return ids.value, nil }

func serviceFixture(store Store) *Service {
	service, err := NewService(store,
		clockFixture{value: time.Date(2026, 8, 1, 1, 2, 3, 456000, time.UTC)},
		idFixture{value: domain.ID(uuid.MustParse("01900000-0000-7000-8000-000000000001"))})
	if err != nil {
		panic(err)
	}
	return service
}

func accessFixture(scope AccessScope, permission, operation string) Access {
	tenantID, slug, name := uuid.Nil, "", ""
	if scope != ScopePlatform {
		tenantID, slug, name = uuid.MustParse("10000000-0000-4000-8000-000000000001"),
			"tenant-a", "Tenant A"
	}
	access, err := NewAccess(uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		tenantID, 1, 1, 1, scope, permission, operation,
		"Ada", "ada@example.com", slug, name)
	if err != nil {
		panic(err)
	}
	return access
}
