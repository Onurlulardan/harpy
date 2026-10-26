package iam

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

const reauthenticationBodyLimit int64 = 72 * 1024

type reauthenticationStartDocument struct {
	PurposeSHA256   string                         `json:"purpose_sha256"`
	Method          AuthHTTPReauthenticationMethod `json:"method"`
	AuthenticatorID json.RawMessage                `json:"authenticator_id"`
}

type reauthenticationCompleteDocument struct {
	ChallengeID   string                         `json:"challenge_id"`
	Challenge     string                         `json:"challenge"`
	PurposeSHA256 string                         `json:"purpose_sha256"`
	Method        AuthHTTPReauthenticationMethod `json:"method"`
	Proof         json.RawMessage                `json:"proof"`
}

func decodeReauthenticationStart(
	w http.ResponseWriter, r *http.Request, origin string,
) (AuthHTTPReauthenticationStartInput, error) {
	var document reauthenticationStartDocument
	if err := httpinput.DecodeJSON(w, r, 2048, &document); err != nil {
		return AuthHTTPReauthenticationStartInput{}, err
	}
	purpose, err := decodeReauthenticationDigest(document.PurposeSHA256)
	if err != nil {
		return AuthHTTPReauthenticationStartInput{}, httpinput.ErrInvalid
	}
	authenticatorID, valid := decodeReauthenticationAuthenticator(
		document.Method, document.AuthenticatorID,
	)
	if !valid {
		return AuthHTTPReauthenticationStartInput{}, httpinput.ErrInvalid
	}
	return AuthHTTPReauthenticationStartInput{Method: document.Method,
		AuthenticatorID: authenticatorID, PurposeSHA256: purpose, Origin: origin}, nil
}

func decodeReauthenticationComplete(
	w http.ResponseWriter, r *http.Request, origin string,
) (AuthHTTPReauthenticationCompleteInput, error) {
	var document reauthenticationCompleteDocument
	if err := httpinput.DecodeJSON(w, r, reauthenticationBodyLimit, &document); err != nil {
		return AuthHTTPReauthenticationCompleteInput{}, err
	}
	challengeID, err := uuid.Parse(document.ChallengeID)
	if err != nil || challengeID.String() != document.ChallengeID {
		return AuthHTTPReauthenticationCompleteInput{}, httpinput.ErrInvalid
	}
	purpose, err := decodeReauthenticationDigest(document.PurposeSHA256)
	if err != nil || !canonicalBase64URL(document.Challenge, 32) {
		return AuthHTTPReauthenticationCompleteInput{}, httpinput.ErrInvalid
	}
	input := AuthHTTPReauthenticationCompleteInput{ChallengeID: challengeID,
		Challenge: []byte(document.Challenge), Method: document.Method,
		PurposeSHA256: purpose, Origin: origin}
	if err := decodeReauthenticationProof(document.Proof, &input); err != nil {
		clearReauthenticationCompleteInput(&input)
		return AuthHTTPReauthenticationCompleteInput{}, httpinput.ErrInvalid
	}
	return input, nil
}

func decodeReauthenticationDigest(value string) ([32]byte, error) {
	var digest [32]byte
	if !canonicalBase64URL(value, len(digest)) {
		return digest, fmt.Errorf("digest is invalid")
	}
	decoded, _ := base64.RawURLEncoding.Strict().DecodeString(value)
	copy(digest[:], decoded)
	clear(decoded)
	if digest == [32]byte{} {
		return [32]byte{}, fmt.Errorf("digest is empty")
	}
	return digest, nil
}

func canonicalBase64URL(value string, size int) bool {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	valid := err == nil && len(decoded) == size &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	clear(decoded)
	return valid
}

func exactJSON(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("proof has trailing data")
	}
	return nil
}

func validUTF8Secret(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value)
}
