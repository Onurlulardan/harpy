package supporthttp

import (
	"context"
	"crypto/sha256"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

type accessFixture struct{ last AccessRequest }

func (fixture *accessFixture) Verify(_ context.Context,
	request AccessRequest,
) (application.Access, error) {
	fixture.last = request
	tenantID, slug, name := uuid.Nil, "", ""
	if request.Scope != application.ScopePlatform {
		tenantID = uuid.MustParse("10000000-0000-4000-8000-000000000001")
		slug, name = "tenant-a", "Tenant A"
	}
	return application.NewAccess(uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		tenantID, 1, 1, 1, request.Scope, request.Permission, request.Operation,
		"Ada", "ada@example.com", slug, name)
}

type rateFixture struct{}

func (rateFixture) Resolve(*http.Request) ([32]byte, error) {
	return sha256.Sum256([]byte("support-rate")), nil
}

type applicationFixture struct{ contactCalls int }

func (fixture *applicationFixture) SubmitContact(context.Context, domain.ContactInput,
	[32]byte, string, [32]byte,
) error {
	fixture.contactCalls++
	return nil
}
func (*applicationFixture) ListContacts(context.Context, application.Access, domain.ContactFilter) (domain.ContactPage, error) {
	return domain.ContactPage{}, nil
}
func (*applicationFixture) ReadContactStats(context.Context, application.Access) (domain.ContactStats, error) {
	return domain.ContactStats{}, nil
}
func (*applicationFixture) GetContact(context.Context, application.Access, domain.ID) (domain.ContactMessage, error) {
	return domain.ContactMessage{}, nil
}
func (*applicationFixture) UpdateContact(context.Context, application.Access, domain.ContactUpdate) (domain.ContactMessage, error) {
	return domain.ContactMessage{}, nil
}
func (*applicationFixture) CreateBug(context.Context, application.Access, domain.BugInput, string, [32]byte) (domain.BugReport, error) {
	return domain.BugReport{}, nil
}
func (*applicationFixture) ListMyBugs(context.Context, application.Access, domain.BugFilter) (domain.BugPage, error) {
	return domain.BugPage{}, nil
}
func (*applicationFixture) GetMyBug(context.Context, application.Access, domain.ID) (domain.BugReport, error) {
	return domain.BugReport{}, nil
}
func (*applicationFixture) ListTenantBugs(context.Context, application.Access, domain.BugFilter) (domain.BugPage, error) {
	return domain.BugPage{}, nil
}
func (*applicationFixture) ReadTenantBugStats(context.Context, application.Access) (domain.BugStats, error) {
	return domain.BugStats{}, nil
}
func (*applicationFixture) GetTenantBug(context.Context, application.Access, domain.ID) (domain.BugReport, error) {
	return domain.BugReport{}, nil
}
func (*applicationFixture) UpdateTenantBug(context.Context, application.Access, domain.BugUpdate) (domain.BugReport, error) {
	return domain.BugReport{}, nil
}
func (*applicationFixture) ListPlatformBugs(context.Context, application.Access, domain.BugFilter) (domain.BugPage, error) {
	return domain.BugPage{}, nil
}
func (*applicationFixture) ReadPlatformBugStats(context.Context, application.Access) (domain.BugStats, error) {
	return domain.BugStats{}, nil
}
func (*applicationFixture) GetPlatformBug(context.Context, application.Access, domain.ID, domain.ID) (domain.BugReport, error) {
	return domain.BugReport{}, nil
}
func (*applicationFixture) UpdatePlatformBug(context.Context, application.Access, domain.BugUpdate) (domain.BugReport, error) {
	return domain.BugReport{}, nil
}

func handlerFixture() (*Handler, *applicationFixture, *accessFixture) {
	app, access := &applicationFixture{}, &accessFixture{}
	apps := Applications{Contact: app, Reporter: app, Tenant: app, Platform: app}
	handler, err := NewHandler(Dependencies{Access: access, Applications: apps,
		RateKeys: rateFixture{}, Booking: BookingConfig{}, Report: func(error) {},
		Correlation: func(*http.Request) string { return "00000000-0000-4000-8000-000000000001" }})
	if err != nil {
		panic(err)
	}
	return handler, app, access
}
