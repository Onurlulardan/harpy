package formrecordarchive

import (
	"context"
	"testing"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type accessFixture struct {
	err               error
	calls             int
	tenantID, actorID domain.ID
	revision          uint64
	itemIDs           []domain.ID
}

func (fixture *accessFixture) Refresh(_ context.Context, tenantID,
	actorID foundation.ID, revision uint64, itemIDs []foundation.ID,
) error {
	fixture.calls++
	fixture.tenantID, fixture.actorID, fixture.revision = tenantID, actorID, revision
	fixture.itemIDs = append([]domain.ID(nil), itemIDs...)
	return fixture.err
}

type generationFixture struct{ err error }

func (fixture generationFixture) ResolveTenantBillingGenerations(
	context.Context, uuid.UUID,
) (int64, int64, int64, error) {
	return 3, 5, 7, fixture.err
}

func (fixture generationFixture) ResolveTenantDocumentsGenerations(
	context.Context, uuid.UUID,
) (int64, int64, error) {
	return 11, 13, fixture.err
}

type scopeFixture struct{ calls int }

func (fixture *scopeFixture) InTenant(ctx context.Context, _ platformtx.ScopeRequest,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	fixture.calls++
	return action(ctx, platformtx.TenantScope{})
}

type linksFixture struct {
	err   error
	calls int
}

func (fixture *linksFixture) Match(context.Context, platformtx.TenantScope,
	domain.ID, domain.ID, domain.ID, domain.ID, []domain.ID,
) error {
	fixture.calls++
	return fixture.err
}

type archivesFixture struct {
	actorID, tenantID domain.ID
	createRequests    []archiveapp.CreateRequest
	getResult         archiveapp.Snapshot
	getCalls          int
}

func (fixture *archivesFixture) Create(_ context.Context, _ platformtx.TenantScope,
	request archiveapp.CreateRequest,
) (archiveapp.Snapshot, error) {
	fixture.createRequests = append(fixture.createRequests, request)
	digest, valid := archiveapp.RequestDigest(request)
	if !valid {
		return archiveapp.Snapshot{}, archiveapp.ErrInvalid
	}
	return archiveapp.Snapshot{ID: fixtureDomainID("018f0000-0000-7000-8000-000000000131"),
		TenantID: fixture.tenantID, ActorID: fixture.actorID,
		JobID:            fixtureDomainID("018f0000-0000-7000-8000-000000000132"),
		RetryOfArchiveID: request.RetryOfArchiveID, State: domain.ArchivePending,
		RequestSHA256: digest}, nil
}

func (fixture *archivesFixture) Get(context.Context, platformtx.TenantScope,
	archiveapp.GetRequest,
) (archiveapp.Snapshot, error) {
	fixture.getCalls++
	return fixture.getResult, nil
}

type applicationFixture struct {
	application *Application
	access      *accessFixture
	scopes      *scopeFixture
	links       *linksFixture
	archives    *archivesFixture
	actorID     domain.ID
	tenantID    domain.ID
}

func newApplicationFixture(t *testing.T) applicationFixture {
	t.Helper()
	actorID := fixtureDomainID("018f0000-0000-7000-8000-000000000141")
	tenantID := fixtureDomainID("018f0000-0000-7000-8000-000000000142")
	access, scopes := &accessFixture{}, &scopeFixture{}
	links := &linksFixture{}
	archives := &archivesFixture{actorID: actorID, tenantID: tenantID}
	application, err := NewApplication(Dependencies{Access: access,
		Generations: generationFixture{}, Scopes: scopes, Links: links,
		Archives: archives})
	if err != nil {
		t.Fatal(err)
	}
	return applicationFixture{application: application, access: access,
		scopes: scopes, links: links, archives: archives,
		actorID: actorID, tenantID: tenantID}
}

func (fixture applicationFixture) apiAccess(t *testing.T) formrecordarchiveapi.Access {
	t.Helper()
	access, err := formrecordarchiveapi.NewAccess(
		formrecordarchiveapi.ID(fixture.actorID), formrecordarchiveapi.ID(fixture.tenantID),
		2, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	return access
}

func fixtureDomainID(value string) domain.ID {
	parsed := uuid.MustParse(value)
	return domain.ID(parsed)
}
