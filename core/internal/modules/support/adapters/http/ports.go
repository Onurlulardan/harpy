package supporthttp

import (
	"context"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

type AccessRequest struct {
	AuthorizationValues, TenantIDValues, TenantSlugValues []string
	Scope                                                 ports.AccessScope
	Permission, Operation                                 string
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (ports.Access, error)
}

type RateKeyResolver interface {
	Resolve(*http.Request) ([32]byte, error)
}

type ContactApplication interface {
	SubmitContact(context.Context, domain.ContactInput, [32]byte, string, [32]byte) error
	ListContacts(context.Context, ports.Access, domain.ContactFilter) (domain.ContactPage, error)
	ReadContactStats(context.Context, ports.Access) (domain.ContactStats, error)
	GetContact(context.Context, ports.Access, domain.ID) (domain.ContactMessage, error)
	UpdateContact(context.Context, ports.Access, domain.ContactUpdate) (domain.ContactMessage, error)
}

type ReporterBugApplication interface {
	CreateBug(context.Context, ports.Access, domain.BugInput, string, [32]byte) (domain.BugReport, error)
	ListMyBugs(context.Context, ports.Access, domain.BugFilter) (domain.BugPage, error)
	GetMyBug(context.Context, ports.Access, domain.ID) (domain.BugReport, error)
}

type TenantBugApplication interface {
	ListTenantBugs(context.Context, ports.Access, domain.BugFilter) (domain.BugPage, error)
	ReadTenantBugStats(context.Context, ports.Access) (domain.BugStats, error)
	GetTenantBug(context.Context, ports.Access, domain.ID) (domain.BugReport, error)
	UpdateTenantBug(context.Context, ports.Access, domain.BugUpdate) (domain.BugReport, error)
}

type PlatformBugApplication interface {
	ListPlatformBugs(context.Context, ports.Access, domain.BugFilter) (domain.BugPage, error)
	ReadPlatformBugStats(context.Context, ports.Access) (domain.BugStats, error)
	GetPlatformBug(context.Context, ports.Access, domain.ID, domain.ID) (domain.BugReport, error)
	UpdatePlatformBug(context.Context, ports.Access, domain.BugUpdate) (domain.BugReport, error)
}

type Applications struct {
	Contact  ContactApplication
	Reporter ReporterBugApplication
	Tenant   TenantBugApplication
	Platform PlatformBugApplication
}

func (applications Applications) ready() bool {
	return ports.Available(applications.Contact) && ports.Available(applications.Reporter) &&
		ports.Available(applications.Tenant) && ports.Available(applications.Platform)
}

type ErrorSink func(error)
type CorrelationResolver func(*http.Request) string
