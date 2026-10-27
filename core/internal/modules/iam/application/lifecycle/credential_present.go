package lifecycle

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"

	"github.com/google/uuid"
)

const credentialRequestDomain = "harpy/iam/credential-request/v1"

type PresentedCredential struct {
	selector [16]byte
	secret   [32]byte
	active   bool
}

func ParseCredential(wire []byte) (*PresentedCredential, error) {
	if len(wire) != credentialWireBytes || wire[22] != '.' {
		return nil, ErrInvalid
	}
	value := &PresentedCredential{active: true}
	_, selectorErr := base64.RawURLEncoding.Strict().Decode(value.selector[:], wire[:22])
	_, secretErr := base64.RawURLEncoding.Strict().Decode(value.secret[:], wire[23:])
	canonical := make([]byte, credentialWireBytes)
	base64.RawURLEncoding.Encode(canonical[:22], value.selector[:])
	canonical[22] = '.'
	base64.RawURLEncoding.Encode(canonical[23:], value.secret[:])
	valid := selectorErr == nil && secretErr == nil && bytes.Equal(canonical, wire) &&
		value.selector != [16]byte{} && value.secret != [32]byte{}
	clear(canonical)
	if !valid {
		value.Destroy()
		return nil, ErrInvalid
	}
	return value, nil
}

func (value *PresentedCredential) Selector() ([16]byte, error) {
	if value == nil || !value.active {
		return [16]byte{}, ErrInvalid
	}
	return value.selector, nil
}

func (value *PresentedCredential) Proof(tenantID uuid.UUID,
	record CredentialRecord, keys OneTimeKeyring) (CredentialProof, error) {
	purpose := CredentialPurpose(record.Purpose)
	if value == nil || !value.active || tenantID == uuid.Nil ||
		record.Validate() != nil || record.Selector != value.selector || keys == nil {
		return CredentialProof{}, ErrInvalid
	}
	key, err := keys.HMACKey(record.HashKeyID)
	if err != nil || key == [32]byte{} {
		clear(key[:])
		return CredentialProof{}, ErrInvalid
	}
	digest, err := credentialHMAC(key, tenantID, purpose, value.selector, value.secret)
	clear(key[:])
	binding, bindingErr := CredentialBindingSHA256(tenantID, purpose, value.selector)
	valid := err == nil && bindingErr == nil &&
		subtle.ConstantTimeCompare(digest[:], record.SecretHMAC[:]) == 1 &&
		subtle.ConstantTimeCompare(binding[:], record.BindingSHA256[:]) == 1
	clear(digest[:])
	if !valid {
		return CredentialProof{}, ErrInvalid
	}
	return CredentialProof{Selector: value.selector, HashKeyID: record.HashKeyID,
		SecretHMAC: record.SecretHMAC, BindingSHA256: binding}, nil
}

func (value *PresentedCredential) RequestSHA256(
	tenantID uuid.UUID,
	purpose CredentialPurpose,
	tokenID uuid.UUID,
	payload []byte,
) ([32]byte, error) {
	if value == nil || !value.active || tenantID == uuid.Nil ||
		!purpose.Valid() || tokenID == uuid.Nil || len(payload) == 0 {
		return [32]byte{}, ErrInvalid
	}
	mac := hmac.New(sha256.New, value.secret[:])
	_, _ = mac.Write([]byte(credentialRequestDomain))
	_, _ = mac.Write(tenantID[:])
	_, _ = mac.Write(tokenID[:])
	_, _ = mac.Write([]byte{byte(len(purpose))})
	_, _ = mac.Write([]byte(purpose))
	_, _ = mac.Write(payload)
	var digest [32]byte
	copy(digest[:], mac.Sum(nil))
	return digest, nil
}

func (value *PresentedCredential) Destroy() {
	if value == nil {
		return
	}
	clear(value.selector[:])
	clear(value.secret[:])
	value.active = false
}
