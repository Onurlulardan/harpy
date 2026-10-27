package aiprovidercontextv1

import (
	"crypto/ed25519"
	"encoding/base64"
)

type Signer struct {
	keyID      string
	privateKey ed25519.PrivateKey
}

func NewSigner(keyID string, privateKey []byte) (*Signer, error) {
	if keyID == "" || len(privateKey) != ed25519.PrivateKeySize {
		return nil, ErrInvalidContext
	}
	return &Signer{keyID: keyID, privateKey: append(ed25519.PrivateKey(nil), privateKey...)}, nil
}

func (signer *Signer) Sign(claims Claims) ([]byte, error) {
	if signer == nil || len(signer.privateKey) != ed25519.PrivateKeySize {
		return nil, ErrInvalidContext
	}
	claims.ContractVersion, claims.Issuer, claims.Audience = ContractVersion, Issuer, Audience
	claims.KeyID = signer.keyID
	if !claims.valid() {
		return nil, ErrInvalidContext
	}
	payload, err := claimsBytes(claims)
	if err != nil {
		return nil, err
	}
	signature := ed25519.Sign(signer.privateKey, payload)
	token, err := canonicalToken(Token{Claims: claims,
		Signature: base64.RawURLEncoding.EncodeToString(signature)})
	clear(signature)
	return token, err
}

func (signer *Signer) Close() {
	if signer != nil {
		clear(signer.privateKey)
	}
}
