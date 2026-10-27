package audithttp

import (
	"context"
	"crypto/sha256"
	"net/http"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	"github.com/google/uuid"
)

var (
	httpActor  = uuid.MustParse("019c0000-0000-7000-8000-000000000201")
	httpTenant = uuid.MustParse("019c0000-0000-7000-8000-000000000202")
	httpEntry  = uuid.MustParse("019c0000-0000-7000-8000-000000000203")
	httpOwner  = uuid.MustParse("019c0000-0000-7000-8000-000000000204")
	httpNow    = time.Date(2026, 8, 1, 15, 0, 0, 123000, time.UTC)
)

type verifierFixture struct{ err error }

func (fixture verifierFixture) Verify(
	_ context.Context, request AccessRequest,
) (Access, error) {
	if fixture.err != nil {
		return Access{}, fixture.err
	}
	return NewAccess(httpActor, httpTenant, 2, 3, 4,
		request.Permission, request.Operation)
}

type applicationFixture struct {
	filter domain.Filter
	lists  int
	entry  domain.Entry
}

func (fixture *applicationFixture) List(
	_ context.Context, _ Access, filter domain.Filter, _ string,
) (domain.Page, error) {
	fixture.lists++
	fixture.filter = filter
	return domain.Page{Entries: []domain.Entry{fixture.entry}, Page: filter.Page,
		Limit: filter.Limit, Total: 1, TotalPages: 1}, nil
}

func (fixture *applicationFixture) Stats(context.Context, Access) (domain.Stats, error) {
	return domain.Stats{Total: 1, Last24Hours: 1, LastHour: 1,
		ByAction: map[string]int64{"update": 1},
		ByEntity: map[string]int64{"form_record": 1},
		Actors:   []domain.ActorStat{{ActorID: domain.ID(httpActor), Count: 1}},
		Entities: []domain.EntityStat{{EntityType: "form_record",
			EntityID: "record-1", FormSlug: "orders", Count: 1}}}, nil
}

func (fixture *applicationFixture) Get(
	context.Context, Access, uuid.UUID,
) (domain.Entry, error) {
	return fixture.entry, nil
}

func fixtureEntry(classification domain.Classification) domain.Entry {
	before, after := "old", "new"
	change := domain.FieldChange{Field: "status", Classification: classification,
		BeforeSHA256: sha256.Sum256([]byte(before)), AfterSHA256: sha256.Sum256([]byte(after))}
	if classification == domain.ClassificationPublic || classification == domain.ClassificationInternal {
		change.Before, change.After = &before, &after
	}
	return domain.Entry{ID: domain.ID(httpEntry), OwnerAuditID: domain.ID(httpOwner),
		ActorID:     domain.ID(httpActor),
		OwnerModule: "forms", Operation: "forms.record.update",
		EntityType: "form_record", EntityID: "record-1", FormSlug: "orders",
		Action: domain.ActionUpdate, AuthorizationVersion: 3,
		FieldChanges: []domain.FieldChange{change}, EventSHA256: sha256.Sum256([]byte("event")),
		OccurredAt: httpNow, RecordedAt: httpNow}
}

func fixtureHandler(t *testing.T, app *applicationFixture) *Handler {
	t.Helper()
	handler, err := NewHandler(Dependencies{Access: verifierFixture{}, Application: app,
		Report: func(error) {}, Correlation: func(*http.Request) string { return "corr-safe" }})
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
