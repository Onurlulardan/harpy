package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPermissionRequestBulkCreateUsesOneCanonicalGroupCommand(t *testing.T) {
	fixture := newPermissionRequestHTTPTestFixture(t)
	body := `{
  "mode":"SELF_SERVICE",
  "expires_at":"2033-05-18T04:34:00Z",
  "items":[
    {
      "target_kind":"USER",
      "target_id":"10000000-0000-4000-8000-000000000001",
      "permission_id":"60000000-0000-4000-8000-000000000001"
    },
    {
      "target_kind":"USER",
      "target_id":"10000000-0000-4000-8000-000000000001",
      "permission_id":"60000000-0000-4000-8000-000000000002"
    }
  ]
}`
	request := fixture.request(
		http.MethodPost, "/api/permission-requests/bulk", body,
	)
	request.Header.Set("Idempotency-Key", "permission-bulk-0001")
	response := httptest.NewRecorder()

	fixture.handler.bulkCreate(response, request)

	if response.Code != http.StatusCreated || fixture.app.createCalls != 1 {
		t.Fatalf("bulk create = %d / calls=%d", response.Code, fixture.app.createCalls)
	}
	input := fixture.app.createInput
	if len(input.Items) != 2 || input.IdempotencyKey != "permission-bulk-0001" ||
		input.Items[0].PermissionID == input.Items[1].PermissionID {
		t.Fatalf("bulk command = %#v", input)
	}
}
