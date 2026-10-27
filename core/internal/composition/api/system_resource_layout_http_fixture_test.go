package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

const layoutHTTPFixtureID = "018f47f4-7b1b-7cc0-8000-000000000033"

type layoutHTTPAccessFixture struct {
	admission layouts.Admission
	request   layouts.LayoutAccessRequest
	err       error
}

func (fixture *layoutHTTPAccessFixture) Verify(
	_ context.Context, request layouts.LayoutAccessRequest,
) (layouts.LayoutAccessDecision, error) {
	fixture.request = request
	if fixture.err != nil {
		return layouts.LayoutAccessDecision{}, fixture.err
	}
	return layouts.NewLayoutAccessDecision(fixture.admission)
}

type layoutHTTPApplicationFixture struct {
	revision                                      layouts.Revision
	lastID                                        string
	lastExpected                                  int64
	createCalls, getCalls, listCalls, updateCalls int
	deleteCalls, publishCalls, resolveCalls       int
	err                                           error
}

func (fixture *layoutHTTPApplicationFixture) Create(
	_ context.Context, _ layouts.Admission, _ layouts.Resource,
	id string, _ layouts.Definition,
) (layouts.Revision, error) {
	fixture.createCalls++
	fixture.lastID = id
	return fixture.revision, fixture.err
}

func (fixture *layoutHTTPApplicationFixture) Get(
	_ context.Context, _ layouts.Admission, _ layouts.Resource,
	id string,
) (layouts.Revision, error) {
	fixture.getCalls++
	fixture.lastID = id
	return fixture.revision, fixture.err
}

func (fixture *layoutHTTPApplicationFixture) List(
	context.Context, layouts.Admission, layouts.Resource,
) ([]layouts.Revision, error) {
	fixture.listCalls++
	return []layouts.Revision{fixture.revision}, fixture.err
}

func (fixture *layoutHTTPApplicationFixture) Update(
	_ context.Context, _ layouts.Admission, _ layouts.Resource,
	input layouts.UpdateInput,
) (layouts.Revision, error) {
	fixture.updateCalls++
	fixture.lastID, fixture.lastExpected = input.ID, input.ExpectedRowRevision
	return fixture.revision, fixture.err
}

func (fixture *layoutHTTPApplicationFixture) Delete(
	_ context.Context, _ layouts.Admission, _ layouts.Resource,
	input layouts.DeleteInput,
) error {
	fixture.deleteCalls++
	fixture.lastID, fixture.lastExpected = input.ID, input.ExpectedRowRevision
	return fixture.err
}

func (fixture *layoutHTTPApplicationFixture) Publish(
	_ context.Context, _ layouts.Admission, input layouts.PublishInput,
) (layouts.Revision, error) {
	fixture.publishCalls++
	fixture.lastID, fixture.lastExpected = input.RevisionID, input.ExpectedRowRevision
	return fixture.revision, fixture.err
}

func (fixture *layoutHTTPApplicationFixture) Resolve(
	context.Context, layouts.Admission, layouts.Resource,
) (layouts.PresentationDocument, error) {
	fixture.resolveCalls++
	return layouts.PresentationDocument{
		RegistryContract: layouts.PresentationRegistryContract,
		DefaultLocale:    "en", Locales: []string{"en"}, Surfaces: []layouts.PresentationSurface{},
	}, fixture.err
}

type layoutHTTPFixture struct {
	handler *systemResourceLayoutHTTPHandler
	access  *layoutHTTPAccessFixture
	app     *layoutHTTPApplicationFixture
}

func newLayoutHTTPFixture(t *testing.T) layoutHTTPFixture {
	t.Helper()
	tenant, _ := layouts.ParseID("018f47f4-7b1b-7cc0-8000-000000000001")
	actorID, _ := layouts.ParseID("018f47f4-7b1b-7cc0-8000-000000000002")
	actor, _ := layouts.NewActorEvidence(tenant, actorID, 1, 1)
	admission, _ := layouts.NewAdmission(tenant, actor, 7, 7, 7)
	revision := layouts.Revision{
		ID: layoutHTTPFixtureID, Resource: layouts.LayoutResourceUser,
		Version: 1, State: layouts.RevisionDraft,
		Definition: layoutHTTPEmptyDefinition(), RowRevision: 2,
		CreatedAt: time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 18, 8, 1, 0, 0, time.UTC),
	}
	access := &layoutHTTPAccessFixture{admission: admission}
	app := &layoutHTTPApplicationFixture{revision: revision}
	handler, err := newSystemResourceLayoutHTTPHandler(systemResourceLayoutHTTPDependencies{
		Access: access, Application: app, Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	return layoutHTTPFixture{handler: handler, access: access, app: app}
}

func layoutHTTPRequest(t *testing.T, method, body string) *http.Request {
	t.Helper()
	request := httptest.NewRequest(method, "/ignored", strings.NewReader(body))
	request.SetPathValue("resource", "users")
	request.SetPathValue("revision_id", layoutHTTPFixtureID)
	request.Header.Set("Authorization", "Bearer fixture")
	request.Header.Set("X-Tenant-ID", "018f47f4-7b1b-7cc0-8000-000000000001")
	request = request.WithContext(requestmeta.WithCorrelationID(
		request.Context(), "layout-request-1",
	))
	return request
}

func layoutHTTPDefinitionBody(t *testing.T) string {
	t.Helper()
	document, err := json.Marshal(layoutHTTPDefinitionRequest{
		Definition: layoutHTTPDefinition{
			Contract: layouts.DefinitionContractV2,
			Surfaces: layoutHTTPEmptyDefinition().Surfaces,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(document)
}

func layoutHTTPEmptyDefinition() layouts.Definition {
	kinds := []layouts.SurfaceKind{
		layouts.SurfaceRecordCreate, layouts.SurfaceRecordUpdate,
		layouts.SurfaceRecordDetail, layouts.SurfaceRecordList,
		layouts.SurfaceRecordFilter,
	}
	surfaces := make([]layouts.SurfaceDefinition, len(kinds))
	for index, kind := range kinds {
		surfaces[index] = layouts.SurfaceDefinition{
			Kind: kind, Sections: []layouts.Section{},
			Translations: []layouts.Translation{},
		}
	}
	return layouts.Definition{
		SourceContract: layouts.DefinitionContractV2, Surfaces: surfaces,
	}
}
