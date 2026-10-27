package billing

import (
	"errors"
	"io"
	"net/http"
	"strings"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

const maximumWebhookBodyBytes int64 = 1024 * 1024

func (handler *Handler) receiveWebhook(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	if r.URL.RawQuery != "" || !validProvider(provider) ||
		r.ContentLength > maximumWebhookBodyBytes {
		handler.writeFailure(w, r, billingRequestInvalid)
		return
	}
	if !handler.webhook.AcceptsProvider(provider) {
		handler.writeFailure(w, r, billingapi.ErrNotFound)
		return
	}
	reader := http.MaxBytesReader(w, r.Body, maximumWebhookBodyBytes)
	body, err := io.ReadAll(reader)
	if err != nil {
		clear(body)
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			handler.writeFailure(w, r, httpinput.ErrBodyTooLarge)
			return
		}
		handler.writeFailure(w, r, billingRequestInvalid)
		return
	}
	if len(body) == 0 {
		clear(body)
		handler.writeFailure(w, r, billingRequestInvalid)
		return
	}
	defer clear(body)
	result, err := handler.webhook.ReceiveWebhook(r.Context(), billingapi.WebhookRequest{
		Provider: provider, Headers: cloneHeaders(r.Header), Body: body,
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	status := http.StatusAccepted
	if result.Duplicate {
		status = http.StatusOK
	}
	handler.writeSuccess(w, r, status, result)
}

func validProvider(value string) bool {
	if value == "" || len(value) > 63 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value[1:] {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && !strings.ContainsRune("._-", character) {
			return false
		}
	}
	return true
}

func cloneHeaders(header http.Header) map[string][]string {
	cloned := make(map[string][]string, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}
