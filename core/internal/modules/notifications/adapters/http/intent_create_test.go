package notificationhttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const validIntentDocument = `{"notification_type":"notifications.custom.ready",` +
	`"recipient_principal_id":"10000000-0000-4000-8000-000000000002",` +
	`"template_version_id":"10000000-0000-4000-8000-000000000003",` +
	`"provider_config_version_id":"00000000-0000-0000-0000-000000000000",` +
	`"locale":"tr","effect_key":"custom/ready/1","category":"system",` +
	`"channels":["in_app"],"data":[],` +
	`"data_digest_sha256":"e3b0c44298fc1c149afbf4c8996fb924` +
	`27ae41e4649b934ca495991b7852b855",` +
	`"correlation_id":"10000000-0000-4000-8000-000000000004",` +
	`"causation_id":"10000000-0000-4000-8000-000000000005"}`

func TestCreateIntentDerivesPublicProducer(t *testing.T) {
	access, _ := NewAccess(httpTestUser, httpTestTenant, 1, 1,
		PermissionSend, OperationIntentCreate, "", false, false)
	application := &httpApplicationFake{}
	handler := newHTTPTestHandler(access, application)
	request := httptest.NewRequest(http.MethodPost, "/api/notification-intents",
		strings.NewReader(validIntentDocument))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.createIntent(recorder, request)
	if recorder.Code != http.StatusCreated ||
		application.lastIntent.Request.Producer != "notifications" {
		t.Fatalf("status=%d producer=%q", recorder.Code,
			application.lastIntent.Request.Producer)
	}
}

func TestCreateIntentRejectsClaimedInternalProducer(t *testing.T) {
	access, _ := NewAccess(httpTestUser, httpTestTenant, 1, 1,
		PermissionSend, OperationIntentCreate, "", false, false)
	application := &httpApplicationFake{}
	handler := newHTTPTestHandler(access, application)
	spoofed := strings.Replace(validIntentDocument, "{",
		`{"producer":"iam",`, 1)
	request := httptest.NewRequest(http.MethodPost, "/api/notification-intents",
		strings.NewReader(spoofed))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.createIntent(recorder, request)
	if recorder.Code != http.StatusBadRequest ||
		application.lastIntent.Request.Producer != "" {
		t.Fatalf("status=%d producer=%q", recorder.Code,
			application.lastIntent.Request.Producer)
	}
}
