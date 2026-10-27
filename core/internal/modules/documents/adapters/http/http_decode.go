package documentshttp

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const maximumJSONBodyBytes = int64(1 << 20)

var (
	errDocumentInvalid  = errors.New("Documents HTTP document is invalid")
	errBodyTooLarge     = errors.New("Documents HTTP body is too large")
	errMediaUnsupported = errors.New("Documents HTTP media type is unsupported")
	errRequestInvalid   = errors.New("Documents HTTP request is invalid")
)

func decodeDocument(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if w == nil || r == nil || r.Body == nil || !jsonMediaType(r.Header) {
		return nil, errMediaUnsupported
	}
	if r.ContentLength > maximumJSONBodyBytes {
		return nil, errBodyTooLarge
	}
	document, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maximumJSONBodyBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, errBodyTooLarge
		}
		return nil, errDocumentInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, int(maximumJSONBodyBytes))
	clear(document)
	if err != nil {
		return nil, errDocumentInvalid
	}
	return canonical, nil
}

func jsonMediaType(header http.Header) bool {
	if len(header.Values("Content-Encoding")) != 0 || len(header.Values("Content-Type")) != 1 {
		return false
	}
	mediaType, parameters, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" || len(parameters) > 1 {
		return false
	}
	charset, present := parameters["charset"]
	return !present || strings.EqualFold(charset, "utf-8")
}
