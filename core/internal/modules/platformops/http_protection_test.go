package platformops

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestEnableProtectionUsesPurposeBoundAccessAndHeaderIdentity(t *testing.T) {
	handler, access, protection := httpHandlerFixture()
	publicKey := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	body := `{"claim":{"algorithm":"` + platformopsapi.RecoveryClaimAlgorithm +
		`","public_key_base64url":"` + publicKey + `"}}`
	request := httptest.NewRequest(http.MethodPost,
		"/api/platform/recovery-protection/enable", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer fixture")
	request.Header.Set("Idempotency-Key", "recovery-enable-0001")
	recorder := httptest.NewRecorder()
	handler.enableProtection(recorder, request)
	if recorder.Code != http.StatusAccepted ||
		recorder.Header().Get("Location") != "/api/platform/recovery-operations/"+httpOperationID {
		t.Fatalf("enable response=%d %s", recorder.Code, recorder.Body.String())
	}
	if len(access.requests) != 1 || !access.requests[0].RequireStepUp ||
		access.requests[0].Capability != "platform.recovery.enable" {
		t.Fatal("enable did not require exact purpose-bound capability")
	}
	want, _ := platformopsapi.StepUpPurpose(httpTenantID,
		"platform.recovery.enable", httpEnvironmentID)
	if access.requests[0].StepUpPurpose != want || protection.request.Kind != "enable" ||
		protection.request.IdempotencyKey != "recovery-enable-0001" ||
		protection.request.ActorUserID != httpActorID ||
		len(protection.request.ClaimPublicKey) != 32 {
		t.Fatal("enable command binding drifted")
	}
	firstOperationID := protection.request.OperationID
	second := httptest.NewRequest(http.MethodPost,
		"/api/platform/recovery-protection/enable", bytes.NewBufferString(body))
	second.Header = request.Header.Clone()
	handler.enableProtection(httptest.NewRecorder(), second)
	if protection.request.OperationID != firstOperationID {
		t.Fatal("Idempotency-Key did not produce stable operation identity")
	}
}

func TestProtectionRoutesExposeImplementedSurface(t *testing.T) {
	handler, _, _ := httpHandlerFixture()
	if routes := handler.Routes(); len(routes) != 38 {
		t.Fatalf("implemented PlatformOps route count=%d", len(routes))
	}
}
