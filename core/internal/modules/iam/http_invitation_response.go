package iam

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type invitationListEnvelope struct {
	Data       []InvitationView `json:"data"`
	Total      int              `json:"total"`
	NextCursor string           `json:"next_cursor,omitempty"`
}

type invitationMessage struct {
	Message string `json:"message"`
}

func (handler *InvitationHTTP) writeJSON(
	w http.ResponseWriter, r *http.Request, status int, payload any,
) {
	if err := httpwire.WriteJSON(
		w, status, authHTTPCorrelationID(r), payload, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM invitation response: %w", err))
	}
}

func (handler *InvitationHTTP) writeInvitation(
	w http.ResponseWriter, r *http.Request, status int, view InvitationView,
) {
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, view.Revision))
	handler.writeJSON(w, r, status, view)
}
