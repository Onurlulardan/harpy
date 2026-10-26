package iam

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func TestReauthenticationStartPublishesPurposeBoundChallenge(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	purpose := [32]byte{9}
	fixture.app.reauthenticationStartChallenge = &authHTTPTestReauthenticationChallenge{
		id:      uuid.MustParse("60000000-0000-4000-8000-000000000001"),
		method:  AuthHTTPReauthenticationPassword,
		expires: fixture.now.Add(5 * time.Minute),
		wire:    base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{4}, 32)),
	}
	response := httptest.NewRecorder()
	request := fixture.reauthenticationRequest(t, "/api/auth/reauthentication/start",
		map[string]any{"purpose_sha256": encodeReauthenticationDigest(purpose),
			"method": "password_v1", "authenticator_id": nil})
	fixture.handler.startReauthentication(response, request)
	if response.Code != http.StatusCreated ||
		fixture.app.reauthenticationStartCalls != 1 {
		t.Fatalf("status=%d body=%s calls=%d", response.Code, response.Body.String(),
			fixture.app.reauthenticationStartCalls)
	}
	view := decodeAuthResponse(t, response)
	if view["method"] != "password_v1" || view["webauthn"] != nil ||
		view["action_policy_version"] != float64(AuthHTTPReauthenticationPolicy) ||
		fixture.app.reauthenticationStartInput.PurposeSHA256 != purpose {
		t.Fatalf("challenge response=%v input=%#v", view,
			fixture.app.reauthenticationStartInput)
	}
}

func TestReauthenticationCompleteIssuesOnlyExactStepUpGrant(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	purpose := [32]byte{7}
	fixture.app.reauthenticationCompleteGrant = fixture.newGrant(
		t, fixture.stepUpClaims(t, purpose),
	)
	response := httptest.NewRecorder()
	request := fixture.reauthenticationRequest(t, "/api/auth/reauthentication/complete",
		completeReauthenticationBody(purpose))
	fixture.handler.completeReauthentication(response, request)
	if response.Code != http.StatusOK ||
		fixture.app.reauthenticationCompleteCalls != 1 ||
		len(response.Header().Values("Set-Cookie")) != 2 {
		t.Fatalf("status=%d body=%s cookies=%v", response.Code,
			response.Body.String(), response.Header().Values("Set-Cookie"))
	}
	if fixture.app.reauthenticationCompleteInput.PurposeSHA256 != purpose ||
		fixture.app.reauthenticationCompleteInput.Method != AuthHTTPReauthenticationPassword {
		t.Fatal("completion binding drifted")
	}
}

func TestReauthenticationCompleteRejectsBaselineGrant(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	purpose := [32]byte{7}
	fixture.app.reauthenticationCompleteGrant = fixture.newGrant(t, fixture.claims)
	response := httptest.NewRecorder()
	request := fixture.reauthenticationRequest(t, "/api/auth/reauthentication/complete",
		completeReauthenticationBody(purpose))
	fixture.handler.completeReauthentication(response, request)
	if response.Code != http.StatusServiceUnavailable ||
		len(response.Header().Values("Set-Cookie")) != 0 {
		t.Fatalf("baseline status=%d cookies=%v", response.Code,
			response.Header().Values("Set-Cookie"))
	}
}

func (fixture *authHTTPTestFixture) reauthenticationRequest(
	t *testing.T, path string, document any,
) *http.Request {
	t.Helper()
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	token, err := fixture.security.accessSigner.Sign(fixture.claims)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", fixture.origin)
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func (fixture *authHTTPTestFixture) stepUpClaims(
	t *testing.T, purpose [32]byte,
) accesstoken.Claims {
	t.Helper()
	input := accesstoken.ClaimsInput{Subject: fixture.claims.Subject(),
		Session: fixture.claims.Session(), Tenant: fixture.claims.Tenant(),
		Client: fixture.claims.Client(), IssuedAt: fixture.now,
		JTI:                        uuid.MustParse("70000000-0000-4000-8000-000000000001"),
		TokenVersion:               fixture.claims.TokenVersion(),
		AuthorizationVersion:       fixture.claims.AuthorizationVersion(),
		PlatformAuthorizationClock: fixture.claims.PlatformAuthorizationClock(),
		SecurityEpoch:              fixture.claims.SecurityEpoch(), AccessMode: fixture.claims.AccessMode(),
		AuthenticationTime: fixture.now, Assurance: accesstoken.AssuranceStepUp,
		Methods:           []accesstoken.AuthMethod{accesstoken.AuthMethodPassword},
		StepUpPurposeHash: purpose[:]}
	claims, err := accesstoken.NewClaims(fixture.profile, input)
	if err != nil {
		t.Fatal(err)
	}
	return claims
}

func encodeReauthenticationDigest(value [32]byte) string {
	return base64.RawURLEncoding.EncodeToString(value[:])
}

func completeReauthenticationBody(purpose [32]byte) map[string]any {
	return map[string]any{
		"challenge_id":   "60000000-0000-4000-8000-000000000001",
		"challenge":      base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{4}, 32)),
		"purpose_sha256": encodeReauthenticationDigest(purpose),
		"method":         "password_v1", "proof": map[string]any{"password": "valid-password"},
	}
}

type authHTTPTestReauthenticationChallenge struct {
	id       uuid.UUID
	method   AuthHTTPReauthenticationMethod
	expires  time.Time
	wire     string
	revealed bool
}

func (challenge *authHTTPTestReauthenticationChallenge) ID() uuid.UUID { return challenge.id }
func (challenge *authHTTPTestReauthenticationChallenge) Method() AuthHTTPReauthenticationMethod {
	return challenge.method
}
func (challenge *authHTTPTestReauthenticationChallenge) ExpiresAt() time.Time {
	return challenge.expires
}
func (*authHTTPTestReauthenticationChallenge) CredentialID() []byte           { return nil }
func (*authHTTPTestReauthenticationChallenge) RPID() string                   { return "" }
func (*authHTTPTestReauthenticationChallenge) UserVerificationRequired() bool { return false }
func (challenge *authHTTPTestReauthenticationChallenge) RevealOnce(consume func([]byte) error) error {
	if challenge.revealed {
		return errReauthenticationProof
	}
	challenge.revealed = true
	return consume([]byte(challenge.wire))
}
func (challenge *authHTTPTestReauthenticationChallenge) Destroy() { challenge.wire = "" }
