package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

type ApplicationKernel interface {
	ContactSubmit(context.Context, domain.ContactInput, [32]byte, string, [32]byte) error
	ContactsList(context.Context, Access, domain.ContactFilter) (domain.ContactPage, error)
	ContactStatsRead(context.Context, Access) (domain.ContactStats, error)
	ContactGet(context.Context, Access, domain.ID) (domain.ContactMessage, error)
	ContactUpdate(context.Context, Access, domain.ContactUpdate) (domain.ContactMessage, error)
	BugCreate(context.Context, Access, domain.BugInput, string, [32]byte) (domain.BugReport, error)
	ReporterBugsList(context.Context, Access, domain.BugFilter) (domain.BugPage, error)
	ReporterBugGet(context.Context, Access, domain.ID) (domain.BugReport, error)
	TenantBugsList(context.Context, Access, domain.BugFilter) (domain.BugPage, error)
	TenantBugStatsRead(context.Context, Access) (domain.BugStats, error)
	TenantBugGet(context.Context, Access, domain.ID) (domain.BugReport, error)
	TenantBugUpdate(context.Context, Access, domain.BugUpdate) (domain.BugReport, error)
	PlatformBugsList(context.Context, Access, domain.BugFilter) (domain.BugPage, error)
	PlatformBugStatsRead(context.Context, Access) (domain.BugStats, error)
	PlatformBugGet(context.Context, Access, domain.ID, domain.ID) (domain.BugReport, error)
	PlatformBugUpdate(context.Context, Access, domain.BugUpdate) (domain.BugReport, error)
}

type ContactDriver interface {
	ContactSubmit(context.Context, ContactSubmission) (domain.ContactMessage, bool, error)
	ContactsList(context.Context, Access, domain.ContactFilter) (domain.ContactPage, error)
	ContactStatsRead(context.Context, Access, time.Time) (domain.ContactStats, error)
	ContactGet(context.Context, Access, domain.ID) (domain.ContactMessage, error)
	ContactUpdate(context.Context, Access, domain.ContactUpdate) (domain.ContactMessage, bool, error)
}

type BugDriver interface {
	BugCreate(context.Context, BugSubmission) (domain.BugReport, bool, error)
	BugsList(context.Context, Access, domain.BugFilter) (domain.BugPage, error)
	BugStatsRead(context.Context, Access, domain.ID) (domain.BugStats, error)
	BugGet(context.Context, Access, domain.ID, domain.ID) (domain.BugReport, error)
	BugUpdate(context.Context, Access, domain.BugUpdate) (domain.BugReport, bool, error)
}
