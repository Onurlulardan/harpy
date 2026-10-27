package invitation

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"

	"github.com/google/uuid"
)

type PresentedCredential struct {
	selector [16]byte
	secret   [32]byte
	active   bool
}

func ParseCredential(wire []byte) (*PresentedCredential, error) {
	if len(wire) != credentialWireBytes || wire[22] != '.' {
		return nil, ErrCredentialInvalid
	}
	credential := &PresentedCredential{active: true}
	if _, err := base64.RawURLEncoding.Strict().Decode(
		credential.selector[:], wire[:22],
	); err != nil {
		credential.Destroy()
		return nil, ErrCredentialInvalid
	}
	if _, err := base64.RawURLEncoding.Strict().Decode(
		credential.secret[:], wire[23:],
	); err != nil {
		credential.Destroy()
		return nil, ErrCredentialInvalid
	}
	canonical := make([]byte, credentialWireBytes)
	base64.RawURLEncoding.Encode(canonical[:22], credential.selector[:])
	canonical[22] = '.'
	base64.RawURLEncoding.Encode(canonical[23:], credential.secret[:])
	valid := bytes.Equal(canonical, wire) && credential.selector != [16]byte{} &&
		credential.secret != [32]byte{}
	clear(canonical)
	if !valid {
		credential.Destroy()
		return nil, ErrCredentialInvalid
	}
	return credential, nil
}

func (credential *PresentedCredential) Selector() ([16]byte, error) {
	if credential == nil || !credential.active {
		return [16]byte{}, ErrCredentialDestroyed
	}
	return credential.selector, nil
}

func (credential *PresentedCredential) Proof(
	tenantID uuid.UUID,
	tokenID uuid.UUID,
	revision int64,
	keyID string,
	expectedHMAC [32]byte,
	keys OneTimeKeyring,
) (TokenProof, error) {
	if credential == nil || !credential.active || tenantID == uuid.Nil ||
		tokenID == uuid.Nil || revision < 1 || !validTokenKeyID(keyID) ||
		expectedHMAC == [32]byte{} || keys == nil {
		return TokenProof{}, ErrCredentialInvalid
	}
	key, err := keys.HMACKey(keyID)
	if err != nil || key == [32]byte{} {
		clear(key[:])
		return TokenProof{}, ErrCredentialInvalid
	}
	actual, err := invitationSecretHMAC(
		key, tenantID, credential.selector, credential.secret,
	)
	clear(key[:])
	if err != nil || subtle.ConstantTimeCompare(actual[:], expectedHMAC[:]) != 1 {
		clear(actual[:])
		return TokenProof{}, ErrCredentialInvalid
	}
	clear(actual[:])
	binding, err := InvitationBindingSHA256(tenantID, credential.selector)
	if err != nil {
		return TokenProof{}, ErrCredentialInvalid
	}
	return TokenProof{
		ID: tokenID, Selector: credential.selector, HashKeyID: keyID,
		SecretHMAC: expectedHMAC, BindingSHA256: binding, Revision: revision,
	}, nil
}

func (credential *PresentedCredential) Destroy() {
	if credential == nil {
		return
	}
	clear(credential.selector[:])
	clear(credential.secret[:])
	credential.active = false
}
