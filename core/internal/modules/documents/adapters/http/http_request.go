package documentshttp

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func pathID(r *http.Request, name string) (string, bool) {
	if r == nil {
		return "", false
	}
	value := r.PathValue(name)
	id, err := domain.ParseID(value)
	return value, err == nil && id.UUIDv7()
}

func bodyAbsent(r *http.Request) bool {
	return r != nil && (r.Body == nil || r.ContentLength == 0) &&
		len(r.Header.Values("Transfer-Encoding")) == 0
}

func queryCopy(r *http.Request) url.Values {
	if r == nil || r.URL == nil {
		return nil
	}
	return r.URL.Query()
}

func idempotencyKey(r *http.Request) (string, bool) {
	if r == nil || len(r.Header.Values("Idempotency-Key")) != 1 {
		return "", false
	}
	value := r.Header.Get("Idempotency-Key")
	return value, value != "" && len(value) <= 256 && strings.TrimSpace(value) == value
}

func contentRequestValid(r *http.Request) bool {
	return r != nil && r.Body != nil && r.ContentLength > 0 &&
		r.ContentLength <= domain.MaxUploadBytes && len(r.Header.Values("Content-Encoding")) == 0 &&
		len(r.Header.Values("Content-Type")) == 1 && r.URL != nil && r.URL.RawQuery == ""
}
