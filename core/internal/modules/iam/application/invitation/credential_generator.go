package invitation

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/google/uuid"
)

const credentialWireBytes = 66

type OneTimeKeyring interface {
	Active() (string, [32]byte, error)
	HMACKey(string) ([32]byte, error)
}

type CredentialGenerator struct{ randomness io.Reader }

func NewCredentialGenerator() CredentialGenerator {
	return CredentialGenerator{randomness: rand.Reader}
}

type CredentialEvidence struct {
	Selector      [16]byte
	HashKeyID     string
	SecretHMAC    [32]byte
	BindingSHA256 [32]byte
}

type IssuedCredential struct {
	selector [16]byte
	secret   [32]byte
	active   bool
}

func (generator CredentialGenerator) GenerateCredential(
	tenantID uuid.UUID,
	keys OneTimeKeyring,
) (*IssuedCredential, CredentialEvidence, error) {
	if tenantID == uuid.Nil || keys == nil || generator.randomness == nil {
		return nil, CredentialEvidence{}, ErrCredentialInvalid
	}
	keyID, key, err := keys.Active()
	if err != nil || !validTokenKeyID(keyID) || key == [32]byte{} {
		clear(key[:])
		return nil, CredentialEvidence{}, ErrCredentialInvalid
	}
	credential := &IssuedCredential{active: true}
	if _, err = io.ReadFull(generator.randomness, credential.selector[:]); err != nil {
		credential.Destroy()
		clear(key[:])
		return nil, CredentialEvidence{}, ErrCredentialRandom
	}
	if _, err = io.ReadFull(generator.randomness, credential.secret[:]); err != nil ||
		credential.selector == [16]byte{} || credential.secret == [32]byte{} {
		credential.Destroy()
		clear(key[:])
		return nil, CredentialEvidence{}, ErrCredentialRandom
	}
	hmacValue, err := invitationSecretHMAC(
		key, tenantID, credential.selector, credential.secret,
	)
	clear(key[:])
	binding, bindingErr := InvitationBindingSHA256(tenantID, credential.selector)
	if err != nil || bindingErr != nil {
		credential.Destroy()
		return nil, CredentialEvidence{}, ErrCredentialInvalid
	}
	return credential, CredentialEvidence{
		Selector: credential.selector, HashKeyID: keyID,
		SecretHMAC: hmacValue, BindingSHA256: binding,
	}, nil
}

func (credential *IssuedCredential) RevealOnce(consume func([]byte) error) error {
	if credential == nil || !credential.active || consume == nil {
		return ErrCredentialDestroyed
	}
	wire := make([]byte, credentialWireBytes)
	base64.RawURLEncoding.Encode(wire[:22], credential.selector[:])
	wire[22] = '.'
	base64.RawURLEncoding.Encode(wire[23:], credential.secret[:])
	err := consume(wire)
	clear(wire)
	credential.Destroy()
	return err
}

func (credential *IssuedCredential) Destroy() {
	if credential == nil {
		return
	}
	clear(credential.selector[:])
	clear(credential.secret[:])
	credential.active = false
}
