package tenancy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	"github.com/google/uuid"
)

var (
	settingsTestPrincipal = uuid.MustParse("019c0000-0000-7000-8000-000000000001")
	settingsTestTenant    = uuid.MustParse("019c0000-0000-7000-8000-000000000002")
)

type settingsAccessFake struct{ request TenantSettingsAccessRequest }

func (fake *settingsAccessFake) Verify(_ context.Context,
	request TenantSettingsAccessRequest,
) (TenantSettingsAccess, error) {
	fake.request = request
	return NewTenantSettingsAccess(settingsTestPrincipal, settingsTestTenant,
		2, 3, 4, request.Permission, request.Operation)
}

type settingsHTTPApplicationFake struct {
	snapshot settings.Snapshot
	update   settings.UpdateCommand
	test     settings.TestEmailCommand
	called   int
}

func (fake *settingsHTTPApplicationFake) Get(context.Context,
	TenantSettingsAccess,
) (settings.Snapshot, error) {
	fake.called++
	return fake.snapshot, nil
}

func (fake *settingsHTTPApplicationFake) Update(_ context.Context,
	_ TenantSettingsAccess, command settings.UpdateCommand,
) (settings.UpdateResult, error) {
	fake.called++
	fake.update = command
	if command.SMTP != nil {
		copy := *command.SMTP
		copy.Password = append([]byte(nil), command.SMTP.Password...)
		fake.update.SMTP = &copy
	}
	return settings.UpdateResult{Revision: command.ExpectedRevision + 1}, nil
}

func (fake *settingsHTTPApplicationFake) TestEmail(_ context.Context,
	_ TenantSettingsAccess, command settings.TestEmailCommand,
) (settings.TestEmailResult, error) {
	fake.called++
	fake.test = command
	return settings.TestEmailResult{OperationID: uuid.MustParse(
		"019c0000-0000-7000-8000-000000000003")}, nil
}

func TestTenantSettingsHTTPRoutesAndSecretRedaction(t *testing.T) {
	access, application, handler := tenantSettingsHTTPFixture(t)
	routes := handler.Routes()
	if len(routes) != 3 || routes[0].Method != http.MethodGet ||
		routes[0].Pattern != "/api/tenant/settings" ||
		routes[1].Method != http.MethodPut || routes[2].Method != http.MethodPost {
		t.Fatalf("Routes() = %+v", routes)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/tenant/settings", nil)
	response := httptest.NewRecorder()
	routes[0].Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("ETag") != `"7"` {
		t.Fatalf("GET status=%d headers=%v body=%s", response.Code,
			response.Header(), response.Body.String())
	}
	if strings.Contains(response.Body.String(), "smtp-secret") ||
		!strings.Contains(response.Body.String(), `"smtp_password":""`) ||
		!strings.Contains(response.Body.String(), `"smtp_password_configured":true`) {
		t.Fatalf("GET secret redaction body=%s", response.Body.String())
	}
	if application.called != 1 || access.request.Permission != TenantSettingsPermissionRead ||
		access.request.Operation != TenantSettingsOperationGet {
		t.Fatalf("GET access=%+v calls=%d", access.request, application.called)
	}
}

func TestTenantSettingsHTTPRequiresRevisionAndRejectsUnknownFields(t *testing.T) {
	_, application, handler := tenantSettingsHTTPFixture(t)
	missing := tenantSettingsRequest(http.MethodPut, `{"name":"Acme"}`)
	missing.Header.Set("Idempotency-Key", "tenant-settings-0001")
	response := httptest.NewRecorder()
	handler.updateTenantSettings(response, missing)
	if response.Code != http.StatusPreconditionRequired || application.called != 0 {
		t.Fatalf("missing precondition status=%d calls=%d", response.Code, application.called)
	}
	unknown := tenantSettingsRequest(http.MethodPut,
		`{"name":"Acme","unexpected":true}`)
	unknown.Header.Set("Idempotency-Key", "tenant-settings-0001")
	unknown.Header.Set("If-Match", `"7"`)
	response = httptest.NewRecorder()
	handler.updateTenantSettings(response, unknown)
	if response.Code != http.StatusBadRequest || application.called != 0 {
		t.Fatalf("unknown field status=%d calls=%d body=%s", response.Code,
			application.called, response.Body.String())
	}
}

func TestTenantSettingsHTTPForwardsCASIdempotencyAndSMTP(t *testing.T) {
	access, application, handler := tenantSettingsHTTPFixture(t)
	body := `{"name":"Acme Updated","smtp":{"smtp_host":"smtp.example.test",` +
		`"smtp_port":"587","smtp_username":"mailer","smtp_password":"smtp-secret",` +
		`"email_from":"noreply@example.test"}}`
	request := tenantSettingsRequest(http.MethodPut, body)
	request.Header.Set("Idempotency-Key", "tenant-settings-0001")
	request.Header.Set("If-Match", `"7"`)
	response := httptest.NewRecorder()
	handler.updateTenantSettings(response, request)
	if response.Code != http.StatusOK || response.Header().Get("ETag") != `"8"` {
		t.Fatalf("PUT status=%d body=%s", response.Code, response.Body.String())
	}
	if application.update.ExpectedRevision != 7 ||
		application.update.IdempotencyKey != "tenant-settings-0001" ||
		application.update.SMTP == nil || string(application.update.SMTP.Password) != "smtp-secret" {
		t.Fatalf("PUT command=%+v", application.update)
	}
	if access.request.Permission != TenantSettingsPermissionManage ||
		access.request.Operation != TenantSettingsOperationUpdate {
		t.Fatalf("PUT access=%+v", access.request)
	}
}

func TestTenantSettingsHTTPAcceptsIdempotentTestEmail(t *testing.T) {
	_, application, handler := tenantSettingsHTTPFixture(t)
	request := tenantSettingsRequest(http.MethodPost, `{"to":"admin@example.test"}`)
	request.Header.Set("Idempotency-Key", "tenant-test-email-0001")
	request.Header.Set("If-Match", `"7"`)
	response := httptest.NewRecorder()
	handler.testTenantEmail(response, request)
	if response.Code != http.StatusAccepted || application.test.ExpectedRevision != 7 ||
		application.test.To != "admin@example.test" ||
		!strings.Contains(response.Body.String(), `"operation_id"`) {
		t.Fatalf("POST status=%d command=%+v body=%s", response.Code,
			application.test, response.Body.String())
	}
}

func tenantSettingsHTTPFixture(t *testing.T) (
	*settingsAccessFake, *settingsHTTPApplicationFake, *TenantSettingsHTTP,
) {
	t.Helper()
	now := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	application := &settingsHTTPApplicationFake{snapshot: settings.Snapshot{
		RegistrySnapshot: settings.RegistrySnapshot{ID: settingsTestTenant,
			Name: "Acme", Slug: "acme", Status: "ACTIVE", Revision: 7,
			CreatedAt: now, UpdatedAt: now}, Projection: settings.Projection{
			UserCount: 3, FormCount: 4, PendingInvitationCount: 1,
			SMTP: &settings.SMTPView{Host: "smtp.example.test", Port: 587,
				Username: "mailer", From: "noreply@example.test", PasswordConfigured: true}}}}
	access := &settingsAccessFake{}
	handler, err := NewTenantSettingsHTTP(access, application, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	return access, application, handler
}

func tenantSettingsRequest(method, body string) *http.Request {
	request := httptest.NewRequest(method, "/api/tenant/settings", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Tenant-ID", settingsTestTenant.String())
	return request
}

func decodeTenantSettingsBody(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	return result
}
