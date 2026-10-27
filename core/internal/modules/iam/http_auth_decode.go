package iam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

const authLoginBodyLimit int64 = 4096

type authHTTPOperatorLoginDocument struct {
	TenantSlug string          `json:"tenant_slug"`
	Email      string          `json:"email"`
	Password   json.RawMessage `json:"password"`
}

type authHTTPPortalLoginDocument struct {
	TenantSlug string          `json:"tenant_slug"`
	Identifier string          `json:"identifier"`
	Password   json.RawMessage `json:"password"`
}

func decodeOperatorLogin(
	w http.ResponseWriter,
	r *http.Request,
) (AuthHTTPLoginInput, error) {
	var document authHTTPOperatorLoginDocument
	if err := httpinput.DecodeJSON(w, r, authLoginBodyLimit, &document); err != nil {
		return AuthHTTPLoginInput{}, err
	}
	password, err := decodeLoginPassword(document.Password)
	clear(document.Password)
	if err != nil || !validLoginText(document.TenantSlug, 128) ||
		!validLoginText(document.Email, 1024) {
		clear(password)
		return AuthHTTPLoginInput{}, httpinput.ErrInvalid
	}
	return AuthHTTPLoginInput{
		Mode: AuthHTTPLoginOperator, TenantSlug: document.TenantSlug,
		Identifier: document.Email, Password: password,
	}, nil
}

func decodePortalLogin(
	w http.ResponseWriter,
	r *http.Request,
) (AuthHTTPLoginInput, error) {
	var document authHTTPPortalLoginDocument
	if err := httpinput.DecodeJSON(w, r, authLoginBodyLimit, &document); err != nil {
		return AuthHTTPLoginInput{}, err
	}
	password, err := decodeLoginPassword(document.Password)
	clear(document.Password)
	if err != nil || !validLoginText(document.TenantSlug, 128) ||
		!validLoginText(document.Identifier, 512) {
		clear(password)
		return AuthHTTPLoginInput{}, httpinput.ErrInvalid
	}
	return AuthHTTPLoginInput{
		Mode: AuthHTTPLoginPortal, TenantSlug: document.TenantSlug,
		Identifier: document.Identifier, Password: password,
	}, nil
}

func decodeLoginPassword(raw json.RawMessage) ([]byte, error) {
	if len(raw) < 2 || len(raw) > 2048 || string(raw) == "null" {
		return nil, fmt.Errorf("password document is invalid")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil ||
		value == "" || len(value) > 512 || !utf8.ValidString(value) {
		return nil, fmt.Errorf("password document is invalid")
	}
	return []byte(value), nil
}

func validLoginText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value)
}

func decodeOptionalEmptyObject(w http.ResponseWriter, r *http.Request) error {
	if r == nil || r.Body == nil {
		return httpinput.ErrInvalid
	}
	if r.ContentLength == 0 && len(r.TransferEncoding) == 0 &&
		len(r.Header.Values("Content-Type")) == 0 &&
		len(r.Header.Values("Content-Encoding")) == 0 {
		return nil
	}
	var document struct{}
	return httpinput.DecodeJSON(w, r, 128, &document)
}
