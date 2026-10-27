package httpinput

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

const maximumDecoderBodyBytes = 16 * 1024 * 1024

func DecodeJSON[T any](
	w http.ResponseWriter,
	r *http.Request,
	maximumBytes int64,
	target *T,
) error {
	if w == nil || r == nil || r.Body == nil || target == nil ||
		maximumBytes < 1 || maximumBytes > maximumDecoderBodyBytes {
		return ErrInvalid
	}
	if !isJSONMediaType(r.Header) {
		return ErrMediaTypeUnsupported
	}
	if r.ContentLength > maximumBytes {
		return ErrBodyTooLarge
	}
	reader := http.MaxBytesReader(w, r.Body, maximumBytes)
	document, err := io.ReadAll(reader)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return ErrBodyTooLarge
		}
		return ErrInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, int(maximumBytes))
	clear(document)
	if err != nil {
		return ErrInvalid
	}
	defer clear(canonical)
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return ErrInvalid
	}
	return nil
}

func isJSONMediaType(header http.Header) bool {
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
