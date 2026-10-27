package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

const (
	recordHTTPTestTenant = "018f47f4-7b1b-7cc0-8000-000000000001"
	recordHTTPTestActor  = "018f47f4-7b1b-7cc0-8000-000000000002"
	recordHTTPTestID     = "018f47f4-7b1b-7cc0-8000-000000000003"
)

type recordHTTPAccessFixture struct {
	admission systemresources.Admission
	request   systemresources.LayoutAccessRequest
}

func (fixture *recordHTTPAccessFixture) Verify(
	_ context.Context,
	request systemresources.LayoutAccessRequest,
) (systemresources.LayoutAccessDecision, error) {
	fixture.request = request
	return systemresources.NewLayoutAccessDecision(fixture.admission)
}

type recordHTTPApplicationFixture struct {
	deleteCalls int
	deleted     systemresources.DeleteInput
}

func (*recordHTTPApplicationFixture) Fields(
	context.Context, systemresources.Admission, systemresources.Resource,
) ([]systemResourceRecordHTTPField, error) {
	return nil, systemresources.ErrUnavailable
}

func (*recordHTTPApplicationFixture) PrepareCreate(
	context.Context, systemresources.CreateInput,
) (systemresources.PreparedCreate, error) {
	return nil, systemresources.ErrUnavailable
}

func (*recordHTTPApplicationFixture) Create(
	context.Context, systemresources.Admission, systemresources.PreparedCreate,
) (systemResourceRecordHTTPView, error) {
	return systemResourceRecordHTTPView{}, systemresources.ErrUnavailable
}

func (*recordHTTPApplicationFixture) Get(
	context.Context, systemresources.Admission, systemresources.GetInput,
) (systemResourceRecordHTTPView, error) {
	return systemResourceRecordHTTPView{}, systemresources.ErrUnavailable
}

func (*recordHTTPApplicationFixture) Query(
	context.Context, systemresources.Admission, systemresources.QueryInput,
) (systemResourceRecordHTTPQueryView, error) {
	return systemResourceRecordHTTPQueryView{}, systemresources.ErrUnavailable
}

func (*recordHTTPApplicationFixture) Update(
	context.Context, systemresources.Admission, systemresources.UpdateInput,
) (systemResourceRecordHTTPView, error) {
	return systemResourceRecordHTTPView{}, systemresources.ErrUnavailable
}

func (fixture *recordHTTPApplicationFixture) Delete(
	_ context.Context,
	_ systemresources.Admission,
	input systemresources.DeleteInput,
) error {
	fixture.deleteCalls++
	fixture.deleted = input
	return nil
}

type recordHTTPFixture struct {
	handler *systemResourceRecordHTTPHandler
	access  *recordHTTPAccessFixture
	app     *recordHTTPApplicationFixture
}

func newRecordHTTPFixture(t *testing.T) recordHTTPFixture {
	t.Helper()
	tenant, _ := systemresources.ParseID(recordHTTPTestTenant)
	actorID, _ := systemresources.ParseID(recordHTTPTestActor)
	actor, _ := systemresources.NewActorEvidence(tenant, actorID, 1, 1)
	admission, _ := systemresources.NewAdmission(tenant, actor, 7, 7, 7)
	access := &recordHTTPAccessFixture{admission: admission}
	application := &recordHTTPApplicationFixture{}
	handler, err := newSystemResourceRecordHTTPHandler(
		systemResourceRecordHTTPDependencies{
			Access: access, Application: application, Report: func(error) {},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return recordHTTPFixture{handler: handler, access: access, app: application}
}

func recordHTTPRequest(method, target string) *http.Request {
	request := httptest.NewRequest(method, target, nil)
	request.SetPathValue("id", recordHTTPTestID)
	request.Header.Set("Authorization", "Bearer fixture")
	request.Header.Set("X-Tenant-ID", recordHTTPTestTenant)
	request.Header.Set("X-Tenant-Slug", "harpy")
	return request
}
