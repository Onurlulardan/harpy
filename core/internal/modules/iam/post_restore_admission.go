package iam

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type postRestoreAdmissionStore interface {
	Admission(context.Context) (postrestore.Admission, error)
}

type PostRestoreAdmission struct {
	store  postRestoreAdmissionStore
	report HTTPErrorSink
}

func NewPostRestoreAdmission(
	store postRestoreAdmissionStore, report HTTPErrorSink,
) (*PostRestoreAdmission, error) {
	if store == nil || report == nil {
		return nil, fmt.Errorf("IAM post-restore admission dependencies are required")
	}
	return &PostRestoreAdmission{store: store, report: report}, nil
}

func (admission *PostRestoreAdmission) Wrap(next http.Handler) (http.Handler, error) {
	if admission == nil || admission.store == nil || next == nil {
		return nil, fmt.Errorf("IAM post-restore admission is unavailable")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, err := admission.store.Admission(r.Context())
		if err == nil && postRestoreRequestAllowed(state.State, r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			admission.report(fmt.Errorf("IAM post-restore admission read failed: %w", err))
		}
		correlationID := authHTTPCorrelationID(r)
		if writeErr := httpwire.WriteFailure(w, correlationID,
			httpwire.FailureServiceUnavailable, r.Method == http.MethodHead); writeErr != nil {
			admission.report(fmt.Errorf("write IAM post-restore admission failure: %w", writeErr))
		}
	}), nil
}

func postRestoreRequestAllowed(state, method, path string) bool {
	switch state {
	case "normal", postrestore.StateReactivated:
		return true
	case postrestore.StateReactivationRequired:
		return method == http.MethodPost && (path == "/api/auth/post-restore/challenge" ||
			path == "/api/auth/post-restore/reactivate-owner")
	default:
		return false
	}
}
