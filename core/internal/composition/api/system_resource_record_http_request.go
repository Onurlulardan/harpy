package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	publicresource "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

const systemResourceRecordHTTPBodyLimit = 1 << 20

func decodeSystemResourceRecordDocument(
	w http.ResponseWriter,
	r *http.Request,
) (map[string]json.RawMessage, error) {
	var document map[string]json.RawMessage
	if err := httpinput.DecodeJSON(
		w, r, systemResourceRecordHTTPBodyLimit, &document,
	); err != nil {
		return nil, err
	}
	if len(document) == 0 {
		return nil, errSystemResourceRecordHTTPRequestInvalid
	}
	return document, nil
}

func systemResourceRecordID(r *http.Request) (publicresource.ID, bool) {
	if r == nil {
		return publicresource.ID{}, false
	}
	id, err := publicresource.ParseID(r.PathValue("id"))
	return id, err == nil
}

func systemResourceRecordIdempotencyKey(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	values := r.Header.Values("Idempotency-Key")
	if len(values) != 1 || values[0] == "" {
		return "", false
	}
	return values[0], true
}

func systemResourceRecordExpectedRevision(r *http.Request) (int64, bool, bool) {
	if r == nil {
		return 0, false, false
	}
	values := r.Header.Values("If-Match")
	if len(values) == 0 {
		return 0, false, true
	}
	if len(values) != 1 || !strings.HasPrefix(values[0], `"revision-`) ||
		!strings.HasSuffix(values[0], `"`) {
		return 0, false, false
	}
	digits := strings.TrimSuffix(strings.TrimPrefix(values[0], `"revision-`), `"`)
	if digits == "" || digits[0] == '0' {
		return 0, false, false
	}
	revision, err := strconv.ParseInt(digits, 10, 64)
	return revision, err == nil && revision > 0, false
}

func systemResourceRecordBodyAbsent(r *http.Request) bool {
	if r == nil || r.ContentLength > 0 || len(r.TransferEncoding) > 0 ||
		len(r.Header.Values("Content-Type")) > 0 ||
		len(r.Header.Values("Content-Encoding")) > 0 {
		return false
	}
	if r.Body == nil || r.Body == http.NoBody {
		return true
	}
	contents, err := io.ReadAll(io.LimitReader(r.Body, 1))
	return err == nil && len(contents) == 0
}

func destroySystemResourceRecordDocument(document map[string]json.RawMessage) {
	for key, value := range document {
		clear(value)
		delete(document, key)
	}
}
