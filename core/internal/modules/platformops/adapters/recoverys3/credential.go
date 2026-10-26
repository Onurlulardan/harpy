package recoverys3

import (
	"bytes"
	"encoding/json"
	"io"
)

const (
	CredentialSchema       = "recovery-s3-credential.v1"
	MaximumCredentialBytes = 16 * 1024
)

type Credential struct {
	Schema          string `json:"schema"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token,omitempty"`
}

func ParseCredential(document []byte) (Credential, error) {
	if len(document) < 1 || len(document) > MaximumCredentialBytes {
		return Credential{}, ErrObject
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var credential Credential
	if decoder.Decode(&credential) != nil ||
		decoder.Decode(new(any)) != io.EOF ||
		credential.Schema != CredentialSchema ||
		!validCredentialValue(credential.AccessKeyID) ||
		!validCredentialValue(credential.SecretAccessKey) ||
		credential.SessionToken != "" &&
			!validCredentialValue(credential.SessionToken) {
		return Credential{}, ErrObject
	}
	return credential, nil
}

func validCredentialValue(value string) bool {
	if len(value) < 1 || len(value) > 4096 {
		return false
	}
	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}
	return true
}
