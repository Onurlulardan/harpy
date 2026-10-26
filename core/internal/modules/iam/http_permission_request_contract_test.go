package iam

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

func TestPermissionRequestDecisionRequiresCASAndIdempotencyHeaders(t *testing.T) {
	fixture := newPermissionRequestHTTPTestFixture(t)
	id, taskID := uuid.New(), uuid.New()
	body := `{"expected_state":"IN_REVIEW","review_task_id":"` + taskID.String() + `"}`
	request := fixture.request(http.MethodPost, "/api/permission-requests/"+id.String()+"/approve", body)
	request.SetPathValue("id", id.String())
	request.Header.Set("Idempotency-Key", "permission-decision-0001")
	response := httptest.NewRecorder()
	fixture.handler.approve(response, request)
	if response.Code != http.StatusBadRequest || fixture.app.decisionCalls != 0 {
		t.Fatalf("missing If-Match = %d / calls=%d", response.Code, fixture.app.decisionCalls)
	}
	request = fixture.request(http.MethodPost, "/api/permission-requests/"+id.String()+"/approve", body)
	request.SetPathValue("id", id.String())
	request.Header.Set("Idempotency-Key", "permission-decision-0001")
	request.Header.Set("If-Match", `"revision-2"`)
	response = httptest.NewRecorder()
	fixture.handler.approve(response, request)
	if response.Code != http.StatusOK || fixture.app.decisionCalls != 1 ||
		fixture.app.decisionInput.ReviewTaskID != taskID ||
		fixture.app.decisionInput.ExpectedRevision != 2 {
		t.Fatalf("decision = %d / %#v", response.Code, fixture.app.decisionInput)
	}
}

func TestPermissionRequestFailureEnvelopeIsStableAndRedacted(t *testing.T) {
	fixture := newPermissionRequestHTTPTestFixture(t)
	fixture.app.err = ErrPermissionRequestHTTPUnavailable
	request := fixture.request(http.MethodGet, "/api/permission-requests", "")
	response := httptest.NewRecorder()
	fixture.handler.list(response, request)
	body := response.Body.String()
	if response.Code != http.StatusServiceUnavailable ||
		!strings.Contains(body, `"code":"IAM_UNAVAILABLE"`) ||
		strings.Contains(body, fixture.token) || strings.Contains(body, fixture.app.err.Error()) {
		t.Fatalf("failure = %d %s", response.Code, body)
	}
}

func TestPermissionRequestCreateAuthorizationModeIsExplicit(t *testing.T) {
	if action, valid := permissionRequestCreateAction(permissionrequest.ModeSelfService); !valid || action != "" {
		t.Fatalf("self-service action = %q/%v", action, valid)
	}
	if action, valid := permissionRequestCreateAction(permissionrequest.ModePrivileged); !valid || action != "manage" {
		t.Fatalf("privileged action = %q/%v", action, valid)
	}
}

func TestPermissionRequestStepUpRejectsStaleAndPurposeSubstitution(t *testing.T) {
	auth := newAuthHTTPTestFixture(t)
	groupID := uuid.New()
	purpose := permissionRequestPurpose(auth.claims.Tenant(), groupID)
	claims := permissionRequestStepUpClaims(t, auth, purpose[:], time.Minute)
	if _, valid := permissionRequestStepUp(claims, groupID); !valid {
		t.Fatal("fresh purpose-bound step-up was rejected")
	}
	if _, valid := permissionRequestStepUp(claims, uuid.New()); valid {
		t.Fatal("group purpose substitution was accepted")
	}
	stale := permissionRequestStepUpClaims(t, auth, purpose[:], 11*time.Minute)
	if _, valid := permissionRequestStepUp(stale, groupID); valid {
		t.Fatal("stale step-up was accepted")
	}
}

func permissionRequestStepUpClaims(
	t *testing.T,
	auth *authHTTPTestFixture,
	purpose []byte,
	age time.Duration,
) accesstoken.Claims {
	t.Helper()
	input := accesstoken.ClaimsInput{
		Subject: auth.claims.Subject(), Session: auth.claims.Session(),
		Tenant: auth.claims.Tenant(), Client: accesstoken.ClientOperatorConsole,
		IssuedAt: auth.now, JTI: uuid.New(), TokenVersion: 3,
		AuthorizationVersion: 5, AccessMode: accesstoken.AccessModeOperator,
		PlatformAuthorizationClock: 7, SecurityEpoch: [32]byte{1},
		AuthenticationTime: auth.now.Add(-age), Assurance: accesstoken.AssuranceStepUp,
		Methods:           []accesstoken.AuthMethod{accesstoken.AuthMethodPassword},
		StepUpPurposeHash: append([]byte(nil), purpose...),
	}
	claims, err := accesstoken.NewClaims(auth.profile, input)
	if err != nil {
		t.Fatal(err)
	}
	return claims
}
