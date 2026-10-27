package iam

import (
	"net/http"
	"strings"
	"testing"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

const profileRecoveryPath = "/api/iam/profile-schemas/user/backfills/" +
	"52000000-0000-4000-8000-000000000001/recover"

func TestProfileRecoveryHTTPValidatesTransportContract(t *testing.T) {
	tests := map[string]struct {
		path   string
		body   string
		mutate func(*http.Request)
		status int
	}{
		"missing bearer": {profileRecoveryPath, "", func(r *http.Request) {
			r.Header.Del("Authorization")
		}, http.StatusUnauthorized},
		"bad resource":   {strings.Replace(profileRecoveryPath, "/user/", "/bad/", 1), "", nil, http.StatusBadRequest},
		"bad plan":       {strings.Replace(profileRecoveryPath, "52000000-0000-4000-8000-000000000001", "bad", 1), "", nil, http.StatusBadRequest},
		"missing key":    {profileRecoveryPath, "", nil, http.StatusBadRequest},
		"body forbidden": {profileRecoveryPath, `{}`, nil, http.StatusBadRequest},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newProfileHTTPTestFixture(t)
			request := fixture.request(http.MethodPost, test.path, test.body)
			if name != "missing key" {
				request.Header.Set("Idempotency-Key", "profile-recovery-0001")
			}
			if test.mutate != nil {
				test.mutate(request)
			}
			response := fixture.serve(request)
			if response.Code != test.status || len(fixture.app.calls) != 0 {
				t.Fatalf("status=%d calls=%d body=%s", response.Code,
					len(fixture.app.calls), response.Body)
			}
		})
	}
}

func TestProfileRecoveryHTTPReturnsCreatedReplaySafeEnvelope(t *testing.T) {
	fixture := newProfileHTTPTestFixture(t)
	view := ProfileBackfillRecoveryView{
		PlanID:             uuid.MustParse("52000000-0000-4000-8000-000000000002"),
		RecoveryOfPlanID:   uuid.MustParse("52000000-0000-4000-8000-000000000001"),
		RecoveryRootPlanID: uuid.MustParse("52000000-0000-4000-8000-000000000003"),
		RecoverySequence:   2, Revision: 3, AuthorizationClock: 4,
		State: profilebackfillapp.RecoveryRunning, Replayed: true,
	}
	fixture.app.results["RecoverBackfill"] = view
	request := fixture.request(http.MethodPost, profileRecoveryPath, "")
	request.Header.Set("Idempotency-Key", "profile-recovery-0001")
	response := fixture.serve(request)
	if response.Code != http.StatusCreated ||
		response.Header().Get("ETag") != `"revision-3"` ||
		!strings.Contains(response.Body.String(), view.PlanID.String()) ||
		len(fixture.app.calls) != 1 ||
		fixture.app.calls[0].operation != "RecoverBackfill" {
		t.Fatalf("status=%d etag=%q calls=%+v body=%s", response.Code,
			response.Header().Get("ETag"), fixture.app.calls, response.Body)
	}
}

func TestProfileRecoveryHTTPMapsApplicationConflict(t *testing.T) {
	fixture := newProfileHTTPTestFixture(t)
	fixture.app.failures["RecoverBackfill"] = ErrProfileHTTPConflict
	request := fixture.request(http.MethodPost, profileRecoveryPath, "")
	request.Header.Set("Idempotency-Key", "profile-recovery-0001")
	response := fixture.serve(request)
	if response.Code != http.StatusConflict ||
		!strings.Contains(response.Body.String(), "RESOURCE_CONFLICT") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}
