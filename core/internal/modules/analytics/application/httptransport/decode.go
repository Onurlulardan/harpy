package httptransport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const BodyLimit int64 = 256 << 10

var (
	errDocumentInvalid  = errors.New("analytics HTTP document is invalid")
	errBodyTooLarge     = errors.New("analytics HTTP body is too large")
	errMediaUnsupported = errors.New("analytics HTTP media type is unsupported")
)

func Decode[T any](writer http.ResponseWriter, request *http.Request) (T, error) {
	return DecodeLimit[T](writer, request, BodyLimit)
}

func DecodeLimit[T any](
	writer http.ResponseWriter, request *http.Request, limit int64,
) (T, error) {
	var target T
	if writer == nil || request == nil || request.Body == nil {
		return target, errDocumentInvalid
	}
	if !jsonMediaType(request.Header) {
		return target, errMediaUnsupported
	}
	if request.ContentLength > limit {
		return target, errBodyTooLarge
	}
	document, err := io.ReadAll(http.MaxBytesReader(writer, request.Body, limit))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return target, errBodyTooLarge
		}
		return target, errDocumentInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, int(limit))
	clear(document)
	if err != nil {
		return target, errDocumentInvalid
	}
	defer clear(canonical)
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err = decoder.Decode(&target); err != nil {
		return target, errDocumentInvalid
	}
	return target, nil
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
