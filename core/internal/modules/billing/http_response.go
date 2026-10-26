package billing

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type successEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func (handler *Handler) writeSuccess(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data any,
) {
	if err := httpwire.WriteJSON(
		w, status, correlationID(r), successEnvelope{Success: true, Data: data}, false,
	); err != nil {
		handler.report(fmt.Errorf("write billing HTTP response: %w", err))
	}
}
