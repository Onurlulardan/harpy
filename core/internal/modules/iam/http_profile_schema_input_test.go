package iam

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

var profileHTTPTestSchemaID = uuid.MustParse(
	"41000000-0000-4000-8000-000000000001",
)

func TestProfileHTTPSchemaPublishRejectsUnknownProperty(t *testing.T) {
	fixture := newProfileHTTPTestFixture(t)
	request := fixture.request(
		http.MethodPost, "/api/iam/profile-schemas/user/publish",
		`{"schema_id":"51000000-0000-4000-8000-000000000001",`+
			`"version":2,"expected_active_schema_id":`+
			`"41000000-0000-4000-8000-000000000001",`+
			`"fields":[],"unexpected":true}`,
	)
	request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
	response := fixture.serve(request)
	if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code, len(fixture.app.calls), response.Body)
	}
}
