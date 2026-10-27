package cryptography

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

type RecoveryCode struct {
	secureMaterial
	keyID string
}

func NewRecoveryCode(wire []byte) (*RecoveryCode, error) {
	if len(wire) < 24 || len(wire) > 256 {
		return nil, authentication.ErrAuthenticationContract
	}
	separator := bytes.IndexByte(wire, '.')
	if separator <= 0 || separator == len(wire)-1 ||
		bytes.IndexByte(wire[separator+1:], '.') >= 0 {
		return nil, authentication.ErrAuthenticationContract
	}
	encoded := wire[separator+1:]
	secret := make([]byte, base64.RawURLEncoding.DecodedLen(len(encoded)))
	written, err := base64.RawURLEncoding.Decode(secret, encoded)
	if err != nil || written < 16 || written > 64 {
		clear(secret)
		return nil, authentication.ErrAuthenticationContract
	}
	material, err := newSecureMaterial(secret[:written], 16, 64)
	clear(secret)
	if err != nil {
		return nil, authentication.ErrAuthenticationContract
	}
	return &RecoveryCode{
		secureMaterial: material,
		keyID:          string(wire[:separator]),
	}, nil
}

func (code *RecoveryCode) KeyID() (string, error) {
	if code == nil || !code.valid || code.keyID == "" {
		return "", authentication.ErrAuthenticationContract
	}
	return code.keyID, nil
}

func (code *RecoveryCode) HMAC(key []byte) ([32]byte, error) {
	if len(key) != 32 {
		return [32]byte{}, authentication.ErrAuthenticationContract
	}
	secret, err := code.copy()
	if err != nil {
		return [32]byte{}, authentication.ErrAuthenticationContract
	}
	defer clear(secret)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("harpy:iam-recovery-code:v1\x00"))
	_, _ = mac.Write(secret)
	var result [32]byte
	copy(result[:], mac.Sum(nil))
	return result, nil
}

func (code *RecoveryCode) Destroy() {
	if code == nil {
		return
	}
	code.secureMaterial.Destroy()
	code.keyID = ""
}
