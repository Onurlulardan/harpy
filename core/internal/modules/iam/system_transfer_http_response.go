package iam

import (
	"errors"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type systemTransferSuccessEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

func (handler *SystemTransferHTTP) writeTransferJSON(w http.ResponseWriter,
	r *http.Request, status int, payload any) {
	err := httpwire.WriteJSON(w, status, systemTransferCorrelationID(r),
		systemTransferSuccessEnvelope{Success: true, Data: payload}, false)
	if err != nil {
		handler.report(errors.New("IAM transfer HTTP response write failed"))
	}
}

func systemTransferCorrelationID(r *http.Request) requestmeta.CorrelationID {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value
		}
	}
	return requestmeta.EntropyFailureCorrelationID
}
