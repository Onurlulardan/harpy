package notificationhttp

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

const maximumBodyBytes = 128 * 1024

var (
	errBodyTooLarge     = errors.New("notification HTTP body is too large")
	errMediaUnsupported = errors.New("notification HTTP media type is unsupported")
	errRequestInvalid   = errors.New("notification HTTP request is invalid")
)

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	if r == nil || r.Body == nil || target == nil {
		return errRequestInvalid
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errMediaUnsupported
	}
	if r.ContentLength > maximumBodyBytes {
		return errBodyTooLarge
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(target); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return errBodyTooLarge
		}
		return errRequestInvalid
	}
	if err = decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errRequestInvalid
	}
	return nil
}

func requireEmptyBody(r *http.Request) error {
	if r == nil || r.Body == nil {
		return nil
	}
	var one [1]byte
	read, err := r.Body.Read(one[:])
	if read != 0 || err != nil && !errors.Is(err, io.EOF) {
		return errRequestInvalid
	}
	return nil
}
