package localkey

import (
	"context"
	"crypto/ed25519"
	"fmt"
)

type Signer struct {
	kid        string
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func New(kid string, privateKey ed25519.PrivateKey) (*Signer, error) {
	if kid == "" || len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("local authority signer is invalid")
	}
	keyCopy := append(ed25519.PrivateKey(nil), privateKey...)
	public := append(ed25519.PublicKey(nil), privateKey.Public().(ed25519.PublicKey)...)
	return &Signer{kid: kid, privateKey: keyCopy, publicKey: public}, nil
}

func (signer *Signer) KID() string { return signer.kid }

func (signer *Signer) PublicKey(context.Context) (ed25519.PublicKey, error) {
	return append(ed25519.PublicKey(nil), signer.publicKey...), nil
}

func (signer *Signer) Sign(_ context.Context, message []byte) ([]byte, error) {
	if len(message) == 0 {
		return nil, fmt.Errorf("authority signing input is empty")
	}
	return ed25519.Sign(signer.privateKey, message), nil
}
