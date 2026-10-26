package iam

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/google/uuid"
)

type portalUserEnvelope struct {
	Success bool           `json:"success"`
	Data    PortalUserView `json:"data"`
}

func (handler *PortalUserHTTP) writeView(
	w http.ResponseWriter, r *http.Request, status int, view PortalUserView,
) {
	if view.ID == uuid.Nil || view.BindingID == uuid.Nil || view.Revision < 1 {
		handler.writeFailure(w, r, portalUserFailureUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, view.Revision))
	if err := httpwire.WriteJSON(w, status, authHTTPCorrelationID(r),
		portalUserEnvelope{Success: true, Data: view}, false); err != nil {
		handler.report(fmt.Errorf("write IAM portal user response: %w", err))
	}
}
