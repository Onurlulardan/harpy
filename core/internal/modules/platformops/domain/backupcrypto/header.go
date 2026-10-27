package backupcrypto

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"regexp"
)

type keyWrap struct {
	Algorithm  string `json:"algorithm"`
	KeyID      string `json:"key_id"`
	Nonce      string `json:"nonce,omitempty"`
	Ciphertext string `json:"ciphertext"`
}

type archiveHeader struct {
	Schema        string  `json:"schema"`
	EnvironmentID string  `json:"environment_id"`
	RecoverySetID string  `json:"recovery_set_id"`
	ComponentID   string  `json:"component_id"`
	ChunkSize     int     `json:"chunk_size"`
	BaseNonce     string  `json:"base_nonce"`
	DRWrap        keyWrap `json:"dr_kek_wrap"`
	RecoveryWrap  keyWrap `json:"recovery_hpke_wrap"`
}

var (
	componentPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	uuidPattern      = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
)

func canonicalHeader(value archiveHeader) ([]byte, error) {
	return canonicalObject(value, MaximumHeaderSize)
}

func parseHeader(document []byte) (archiveHeader, error) {
	if len(document) == 0 || len(document) > MaximumHeaderSize {
		return archiveHeader{}, ErrInvalidArchive
	}
	var value archiveHeader
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || decoder.Decode(new(any)) != io.EOF || !value.valid() {
		return archiveHeader{}, ErrInvalidArchive
	}
	canonical, err := canonicalHeader(value)
	if err != nil || !bytes.Equal(canonical, document) {
		return archiveHeader{}, ErrInvalidArchive
	}
	return value, nil
}

func (value archiveHeader) valid() bool {
	return value.Schema == HeaderSchema && validUUID(value.EnvironmentID) &&
		validUUID(value.RecoverySetID) && componentPattern.MatchString(value.ComponentID) &&
		value.ChunkSize == ChunkSize && validEncodedSize(value.BaseNonce, 12) &&
		value.DRWrap.Algorithm == "AES-256-GCM" && validKeyID(value.DRWrap.KeyID) &&
		validEncodedSize(value.DRWrap.Nonce, 12) && validEncoded(value.DRWrap.Ciphertext) &&
		value.RecoveryWrap.Algorithm == "HPKE-X25519-HKDF-SHA256-AES256GCM" &&
		validKeyID(value.RecoveryWrap.KeyID) && value.RecoveryWrap.Nonce == "" &&
		validEncoded(value.RecoveryWrap.Ciphertext)
}

func validUUID(value string) bool {
	return uuidPattern.MatchString(value)
}

func validKeyID(value string) bool { return componentPattern.MatchString(value) }

func validEncoded(value string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) > 0
}

func validEncodedSize(value string, size int) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) == size
}
