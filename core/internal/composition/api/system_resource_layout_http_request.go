package api

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
)

func layoutHTTPResource(r *http.Request) (layouts.Resource, bool) {
	if r == nil {
		return "", false
	}
	return layouts.ParseResource(r.PathValue("resource"))
}

func layoutHTTPRevisionID(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	return layouts.ParseRevisionID(r.PathValue("revision_id"))
}

func layoutHTTPExpectedRevision(r *http.Request) (int64, bool, bool) {
	if r == nil {
		return 0, false, false
	}
	values := r.Header.Values("If-Match")
	if len(values) == 0 {
		return 0, false, true
	}
	if len(values) != 1 || !strings.HasPrefix(values[0], `"layout-row-revision-`) ||
		!strings.HasSuffix(values[0], `"`) {
		return 0, false, false
	}
	digits := strings.TrimSuffix(strings.TrimPrefix(
		values[0], `"layout-row-revision-`), `"`)
	if digits == "" || digits[0] == '0' {
		return 0, false, false
	}
	value, err := strconv.ParseInt(digits, 10, 64)
	return value, err == nil && value > 0, false
}

func layoutHTTPQueryAbsent(r *http.Request) bool {
	return r != nil && r.URL != nil && r.URL.RawQuery == ""
}

func layoutHTTPBodyAbsent(r *http.Request) bool {
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
