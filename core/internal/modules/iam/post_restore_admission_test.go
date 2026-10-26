package iam

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

type postRestoreAdmissionFixture struct {
	state postrestore.Admission
	err   error
}

func (fixture postRestoreAdmissionFixture) Admission(context.Context) (postrestore.Admission, error) {
	return fixture.state, fixture.err
}

func TestPostRestoreAdmissionAllowsOnlyExactRecoveryRoutes(t *testing.T) {
	tests := []struct {
		state, method, path string
		allowed             bool
	}{
		{"normal", http.MethodGet, "/api/forms", true},
		{postrestore.StateReactivated, http.MethodGet, "/api/forms", true},
		{postrestore.StateReactivationRequired, http.MethodPost,
			"/api/auth/post-restore/challenge", true},
		{postrestore.StateReactivationRequired, http.MethodPost,
			"/api/auth/post-restore/reactivate-owner", true},
		{postrestore.StateReactivationRequired, http.MethodGet,
			"/api/auth/post-restore/challenge", false},
		{postrestore.StateReactivationRequired, http.MethodPost,
			"/api/auth/post-restore/challenge/", false},
		{"", http.MethodGet, "/api/forms", false},
		{postrestore.StateFenced, http.MethodPost,
			"/api/auth/post-restore/challenge", false},
	}
	for _, test := range tests {
		t.Run(test.state+test.method+test.path, func(t *testing.T) {
			called := false
			admission, err := NewPostRestoreAdmission(postRestoreAdmissionFixture{
				state: postrestore.Admission{State: test.state}}, func(error) {})
			if err != nil {
				t.Fatal(err)
			}
			handler, err := admission.Wrap(http.HandlerFunc(func(
				w http.ResponseWriter, _ *http.Request,
			) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			}))
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
			if called != test.allowed {
				t.Fatalf("called=%t want=%t status=%d", called, test.allowed, response.Code)
			}
			if !test.allowed && response.Code != http.StatusServiceUnavailable {
				t.Fatalf("denied status=%d", response.Code)
			}
		})
	}
}

func TestPostRestoreAdmissionFailsClosedOnReadError(t *testing.T) {
	reported := false
	admission, err := NewPostRestoreAdmission(postRestoreAdmissionFixture{
		err: errors.New("database unavailable")}, func(error) { reported = true })
	if err != nil {
		t.Fatal(err)
	}
	handler, _ := admission.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler was called")
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/forms", nil))
	if response.Code != http.StatusServiceUnavailable || !reported {
		t.Fatalf("status=%d reported=%t", response.Code, reported)
	}
}
