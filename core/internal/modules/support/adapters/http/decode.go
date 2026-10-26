package supporthttp

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
)

const maximumBodyBytes = int64(64 << 10)

var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,199}$`)

func decodeObject(response http.ResponseWriter, request *http.Request, target any) ([]byte, error) {
	if response == nil || request == nil || request.Body == nil || !jsonMediaType(request.Header) ||
		request.ContentLength > maximumBodyBytes {
		return nil, applicationInvalid()
	}
	document, err := io.ReadAll(http.MaxBytesReader(response, request.Body, maximumBodyBytes))
	if err != nil {
		return nil, applicationInvalid()
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, int(maximumBodyBytes))
	clear(document)
	if err != nil || strictUnmarshal(canonical, target) != nil {
		return nil, applicationInvalid()
	}
	return canonical, nil
}

func jsonMediaType(header http.Header) bool {
	if len(header.Values("Content-Type")) != 1 || len(header.Values("Content-Encoding")) != 0 {
		return false
	}
	mediaType, parameters, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" || len(parameters) > 1 {
		return false
	}
	charset, present := parameters["charset"]
	return !present || strings.EqualFold(charset, "utf-8")
}

func commandEvidence(request *http.Request, canonical []byte) (string, [32]byte, error) {
	if request == nil || len(request.Header.Values("Idempotency-Key")) != 1 {
		return "", [32]byte{}, applicationInvalid()
	}
	key := request.Header.Get("Idempotency-Key")
	if !idempotencyPattern.MatchString(key) {
		return "", [32]byte{}, applicationInvalid()
	}
	return key, sha256.Sum256(canonical), nil
}

func strictUnmarshal(document []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func applicationInvalid() error { return application.ErrInvalid }
