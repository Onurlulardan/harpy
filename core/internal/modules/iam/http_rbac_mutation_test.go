package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestRBACPermissionCreateUsesExactV1TargetWire(t *testing.T) {
	fixture := newRBACHTTPTestFixture(t)
	resourceID, userID, actionID := uuid.New(), uuid.New(), uuid.New()
	body := `{"resource_id":"` + resourceID.String() + `","target":"USER",` +
		`"user_id":"` + userID.String() + `","action_ids":["` + actionID.String() + `"]}`
	request := fixture.request(http.MethodPost, "/api/permissions", body)
	request.Header.Set("Idempotency-Key", rbacHTTPTestKey)
	response := httptest.NewRecorder()
	fixture.handler.createPermission(response, request)
	if response.Code != http.StatusCreated || len(fixture.app.calls) != 1 {
		t.Fatalf("create = %d / calls=%v / %s", response.Code, fixture.app.calls, response.Body.String())
	}
	input := fixture.app.last.(RBACPermissionCreate)
	if input.ResourceID != resourceID || input.TargetID != userID ||
		len(input.ActionIDs) != 1 || input.ActionIDs[0] != actionID {
		t.Fatalf("input = %#v", input)
	}
}

func TestRBACRevisionedDeleteRequiresBothHeaders(t *testing.T) {
	fixture := newRBACHTTPTestFixture(t)
	id := uuid.New()
	for _, missing := range []string{"Idempotency-Key", "If-Match"} {
		request := fixture.request(http.MethodDelete, "/api/permissions/"+id.String(), "")
		request.SetPathValue("id", id.String())
		request.Header.Set("Idempotency-Key", rbacHTTPTestKey)
		request.Header.Set("If-Match", `"revision-1"`)
		request.Header.Del(missing)
		response := httptest.NewRecorder()
		fixture.handler.deletePermission(response, request)
		if response.Code != http.StatusPreconditionRequired || len(fixture.app.calls) != 0 {
			t.Fatalf("missing=%s status=%d calls=%v", missing, response.Code, fixture.app.calls)
		}
	}
}

func TestRBACCatalogCreateDerivesPermanentCompatibilitySlug(t *testing.T) {
	fixture := newRBACHTTPTestFixture(t)
	request := fixture.request(
		http.MethodPost, "/api/permissions/resources",
		`{"name":"Çalışan İzinleri","description":"İK"}`,
	)
	request.Header.Set("Idempotency-Key", rbacHTTPTestKey)
	response := httptest.NewRecorder()
	fixture.handler.createResource(response, request)
	if response.Code != http.StatusCreated || len(fixture.app.calls) != 1 {
		t.Fatalf("create = %d / calls=%v", response.Code, fixture.app.calls)
	}
	input := fixture.app.last.(RBACCatalogCreate)
	if input.Key != "calisan-izinleri" {
		t.Fatalf("slug = %q", input.Key)
	}
}
