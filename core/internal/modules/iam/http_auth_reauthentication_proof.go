package iam

import (
	"encoding/json"
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"
)

var errReauthenticationProof = errors.New("reauthentication proof is invalid")

func decodeReauthenticationAuthenticator(
	method AuthHTTPReauthenticationMethod,
	raw json.RawMessage,
) ([16]byte, bool) {
	if method == AuthHTTPReauthenticationPassword {
		return [16]byte{}, string(raw) == "null"
	}
	if method != AuthHTTPReauthenticationWebAuthn &&
		method != AuthHTTPReauthenticationTOTP {
		return [16]byte{}, false
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return [16]byte{}, false
	}
	id, err := uuid.Parse(value)
	return id, err == nil && id != uuid.Nil && id.String() == value
}

func decodeReauthenticationProof(
	raw json.RawMessage,
	input *AuthHTTPReauthenticationCompleteInput,
) error {
	switch input.Method {
	case AuthHTTPReauthenticationPassword:
		var proof struct {
			Password string `json:"password"`
		}
		if err := exactJSON(raw, &proof); err != nil ||
			!validUTF8Secret(proof.Password, 512) {
			return errReauthenticationProof
		}
		input.Password = []byte(proof.Password)
	case AuthHTTPReauthenticationTOTP:
		var proof struct {
			TOTP string `json:"totp"`
		}
		if err := exactJSON(raw, &proof); err != nil || !sixDigits(proof.TOTP) {
			return errReauthenticationProof
		}
		input.TOTP = []byte(proof.TOTP)
	case AuthHTTPReauthenticationWebAuthn:
		var proof struct {
			Assertion json.RawMessage `json:"webauthn_assertion"`
		}
		if err := exactJSON(raw, &proof); err != nil || len(proof.Assertion) < 2 ||
			len(proof.Assertion) > 64*1024 || proof.Assertion[0] != '{' ||
			!json.Valid(proof.Assertion) {
			return errReauthenticationProof
		}
		input.WebAuthnAssertion = append([]byte(nil), proof.Assertion...)
	default:
		return errReauthenticationProof
	}
	return nil
}

func sixDigits(value string) bool {
	if len(value) != 6 || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
