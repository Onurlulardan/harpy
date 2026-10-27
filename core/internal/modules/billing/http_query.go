package billing

import (
	"encoding/json"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

const maximumQueryBodyBytes int64 = 64 * 1024

func (handler *Handler) decodeGETQuery(r *http.Request) (billingapi.QueryRequest, error) {
	if r == nil || r.URL == nil {
		return billingapi.QueryRequest{}, billingRequestInvalid
	}
	return handler.query.DecodeGET(r.URL.RawQuery)
}

func (handler *Handler) decodePOSTQuery(
	w http.ResponseWriter,
	r *http.Request,
) (billingapi.QueryRequest, error) {
	var document json.RawMessage
	if err := httpinput.DecodeJSON(w, r, maximumQueryBodyBytes, &document); err != nil {
		return billingapi.QueryRequest{}, err
	}
	return handler.query.DecodePOST(document)
}
