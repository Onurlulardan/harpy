package iam

import (
	"encoding/json"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

type authHTTPInitialPasswordDocument struct {
	TenantSlug  string          `json:"tenant_slug"`
	NewPassword json.RawMessage `json:"new_password"`
}

func decodeInitialPassword(
	w http.ResponseWriter,
	r *http.Request,
) (AuthHTTPInitialPasswordInput, error) {
	var document authHTTPInitialPasswordDocument
	if err := httpinput.DecodeJSON(w, r, authLoginBodyLimit, &document); err != nil {
		return AuthHTTPInitialPasswordInput{}, err
	}
	password, err := decodeLoginPassword(document.NewPassword)
	clear(document.NewPassword)
	if err != nil || !validLoginText(document.TenantSlug, 128) {
		clear(password)
		return AuthHTTPInitialPasswordInput{}, httpinput.ErrInvalid
	}
	return AuthHTTPInitialPasswordInput{
		TenantSlug: document.TenantSlug, NewPassword: password,
	}, nil
}
