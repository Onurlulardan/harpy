package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

// ApplicationStore composes resource-specific persistence behind the application port.
type ApplicationStore struct {
	contact  ports.ContactPersistence
	reporter ports.ReporterBugPersistence
	tenant   ports.TenantBugPersistence
	platform ports.PlatformBugPersistence
}

func NewApplicationStore(contact ports.ContactPersistence, reporter ports.ReporterBugPersistence,
	tenant ports.TenantBugPersistence, platform ports.PlatformBugPersistence,
) (*ApplicationStore, error) {
	if !ports.Available(contact) || !ports.Available(reporter) ||
		!ports.Available(tenant) || !ports.Available(platform) {
		return nil, fmt.Errorf("support application persistence is unavailable")
	}
	return &ApplicationStore{contact: contact, reporter: reporter,
		tenant: tenant, platform: platform}, nil
}

func (store *ApplicationStore) ContactSubmit(ctx context.Context,
	command ports.ContactSubmission,
) (domain.ContactMessage, bool, error) {
	return store.contact.SubmitContact(ctx, command)
}

func (store *ApplicationStore) ContactsList(ctx context.Context, access ports.Access,
	filter domain.ContactFilter,
) (domain.ContactPage, error) {
	return store.contact.ListContacts(ctx, access, filter)
}

func (store *ApplicationStore) ContactStatsRead(ctx context.Context, access ports.Access,
	observedAt time.Time,
) (domain.ContactStats, error) {
	return store.contact.ReadContactStats(ctx, access, observedAt)
}

func (store *ApplicationStore) ContactGet(ctx context.Context, access ports.Access,
	id domain.ID,
) (domain.ContactMessage, error) {
	return store.contact.GetContact(ctx, access, id)
}

func (store *ApplicationStore) ContactUpdate(ctx context.Context, access ports.Access,
	update domain.ContactUpdate,
) (domain.ContactMessage, bool, error) {
	return store.contact.UpdateContact(ctx, access, update)
}

func (store *ApplicationStore) BugCreate(ctx context.Context,
	command ports.BugSubmission,
) (domain.BugReport, bool, error) {
	return store.reporter.CreateBug(ctx, command)
}

func (store *ApplicationStore) BugsList(ctx context.Context, access ports.Access,
	filter domain.BugFilter,
) (domain.BugPage, error) {
	switch access.Scope() {
	case ports.ScopeReporter:
		return store.reporter.ListMyBugs(ctx, access, filter)
	case ports.ScopeTenant:
		return store.tenant.ListTenantBugs(ctx, access, filter)
	case ports.ScopePlatform:
		return store.platform.ListPlatformBugs(ctx, access, filter)
	default:
		return domain.BugPage{}, ports.ErrForbidden
	}
}

func (store *ApplicationStore) BugStatsRead(ctx context.Context, access ports.Access,
	tenantID domain.ID,
) (domain.BugStats, error) {
	switch access.Scope() {
	case ports.ScopeTenant:
		return store.tenant.ReadTenantBugStats(ctx, access, tenantID)
	case ports.ScopePlatform:
		return store.platform.ReadPlatformBugStats(ctx, access, tenantID)
	default:
		return domain.BugStats{}, ports.ErrForbidden
	}
}

func (store *ApplicationStore) BugGet(ctx context.Context, access ports.Access,
	id, tenantID domain.ID,
) (domain.BugReport, error) {
	switch access.Scope() {
	case ports.ScopeReporter:
		return store.reporter.GetMyBug(ctx, access, id, tenantID)
	case ports.ScopeTenant:
		return store.tenant.GetTenantBug(ctx, access, id, tenantID)
	case ports.ScopePlatform:
		return store.platform.GetPlatformBug(ctx, access, id, tenantID)
	default:
		return domain.BugReport{}, ports.ErrForbidden
	}
}

func (store *ApplicationStore) BugUpdate(ctx context.Context, access ports.Access,
	update domain.BugUpdate,
) (domain.BugReport, bool, error) {
	switch access.Scope() {
	case ports.ScopeTenant:
		return store.tenant.UpdateTenantBug(ctx, access, update)
	case ports.ScopePlatform:
		return store.platform.UpdatePlatformBug(ctx, access, update)
	default:
		return domain.BugReport{}, false, ports.ErrForbidden
	}
}
