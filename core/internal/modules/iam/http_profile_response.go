package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type profileEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func (handler *ProfileHTTP) writeJSON(
	w http.ResponseWriter, r *http.Request, status int, payload any,
) {
	if err := httpwire.WriteJSON(
		w, status, authHTTPCorrelationID(r), payload, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM profile response: %w", err))
	}
}

func (handler *ProfileHTTP) writeSelf(
	w http.ResponseWriter, r *http.Request, view ProfileSelfView,
) {
	if view.Revision < 1 || !profileSelfDocumentSafe(view.Document) {
		handler.writeFailure(w, r, ErrProfileHTTPUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, view.Revision))
	handler.writeJSON(
		w, r, http.StatusOK,
		profileEnvelope{Success: true, Data: view.Document},
	)
}

func profileSelfDocumentSafe(document []byte) bool {
	var value map[string]json.RawMessage
	if json.Unmarshal(document, &value) != nil || len(value) == 0 {
		return false
	}
	for _, key := range []string{
		"password", "password_hash", "email_normalized", "token_version",
		"authorization_version", "identifier_id", "profile_sha256",
	} {
		if _, exposed := value[key]; exposed {
			return false
		}
	}
	return !bytes.Equal(document, []byte("{}"))
}

func (handler *ProfileHTTP) writeSchema(
	w http.ResponseWriter, r *http.Request, view ProfileSchemaView,
) {
	if view.ID.Version() == 0 || view.Version < 1 {
		handler.writeFailure(w, r, ErrProfileHTTPUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"schema-%s"`, view.ID))
	handler.writeJSON(w, r, http.StatusOK, profileEnvelope{Success: true, Data: view})
}

func (handler *ProfileHTTP) writeMutation(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	view ProfileSchemaMutationView,
) {
	if view.ID.Version() == 0 || view.Revision < 1 || view.AuthorizationClock < 1 {
		handler.writeFailure(w, r, ErrProfileHTTPUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, view.Revision))
	handler.writeJSON(w, r, status, profileEnvelope{Success: true, Data: view})
}

func (handler *ProfileHTTP) writeRecovery(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	view ProfileBackfillRecoveryView,
) {
	if view.PlanID.Version() == 0 || view.RecoveryOfPlanID.Version() == 0 ||
		view.RecoveryRootPlanID.Version() == 0 || view.RecoverySequence < 1 ||
		view.Revision < 1 || view.AuthorizationClock < 1 || !view.State.Valid() {
		handler.writeFailure(w, r, ErrProfileHTTPUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, view.Revision))
	handler.writeJSON(w, r, status, profileEnvelope{Success: true, Data: view})
}
