package lifecyclehttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

func TestPurge(t *testing.T) {
	if err := exerciseInstallation("purge", "catalog.installation.purge",
		func(handler *Handler) Route { return handler.PurgeRoute() }); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeRejectsCallerSuppliedApproverAndStepUp(t *testing.T) {
	access := new(installationAccess)
	application := new(installationApplication)
	runtime, err := cataloghttp.NewRuntime(access, func(error) {}, func(*http.Request) string {
		return "018f0f8b-4c31-7d4e-8a2f-9e0b7c6d5a41"
	})
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]uuid.UUID, 4)
	for index := range ids {
		ids[index], err = uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
	}
	body := fmt.Sprintf(
		`{"operation_id":%q,"backup_receipt_id":%q,"approved_by":%q,`+
			`"step_up_at":"2026-07-30T00:00:00Z","reason":"verified",`+
			`"confirmation_slug":"package","approval_receipt_id":%q}`,
		ids[1], ids[2], ids[3], ids[3],
	)
	request := httptest.NewRequest(http.MethodPost,
		"/api/catalog/installations/"+ids[0].String()+"/purges", strings.NewReader(body))
	request.SetPathValue("installation_id", ids[0].String())
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "catalog-test-key-0001")
	response := httptest.NewRecorder()
	New(runtime, application).PurgeRoute().Handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || application.called != "" ||
		access.operation != "" {
		t.Fatalf("legacy approval body status=%d call=%s operation=%s",
			response.Code, application.called, access.operation)
	}
}
