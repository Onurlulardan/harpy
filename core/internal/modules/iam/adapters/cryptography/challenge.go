package cryptography

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

const challengeBytes = 32

type ChallengeGenerator struct{}

func NewChallengeGenerator() ChallengeGenerator { return ChallengeGenerator{} }

func (ChallengeGenerator) Generate() (authentication.IssuedChallenge, error) {
	raw := make([]byte, challengeBytes)
	if _, err := rand.Read(raw); err != nil {
		return nil, authentication.ErrAuthenticationCrypto
	}
	wire := make([]byte, base64.RawURLEncoding.EncodedLen(len(raw)))
	base64.RawURLEncoding.Encode(wire, raw)
	clear(raw)
	material, err := newSecureMaterial(wire, len(wire), len(wire))
	clear(wire)
	if err != nil {
		return nil, authentication.ErrAuthenticationCrypto
	}
	return &IssuedChallenge{secureMaterial: material}, nil
}

type IssuedChallenge struct {
	secureMaterial
	revealed bool
}

func (challenge *IssuedChallenge) HMAC(key []byte) ([32]byte, error) {
	return challengeHMAC(&challenge.secureMaterial, key)
}

func (challenge *IssuedChallenge) RevealOnce(consume func([]byte) error) error {
	if challenge == nil || challenge.revealed || consume == nil {
		return authentication.ErrAuthenticationContract
	}
	wire, err := challenge.copy()
	if err != nil {
		return authentication.ErrAuthenticationContract
	}
	challenge.revealed = true
	defer clear(wire)
	return consume(wire)
}

type PresentedChallenge struct{ secureMaterial }

func NewPresentedChallenge(wire []byte) (*PresentedChallenge, error) {
	material, err := newSecureMaterial(wire, 43, 43)
	if err != nil {
		return nil, authentication.ErrAuthenticationContract
	}
	decoded := make([]byte, base64.RawURLEncoding.DecodedLen(len(wire)))
	written, err := base64.RawURLEncoding.Decode(decoded, wire)
	if err != nil || written != challengeBytes {
		material.Destroy()
		clear(decoded)
		return nil, authentication.ErrAuthenticationContract
	}
	clear(decoded)
	return &PresentedChallenge{secureMaterial: material}, nil
}

func (challenge *PresentedChallenge) HMAC(key []byte) ([32]byte, error) {
	return challengeHMAC(&challenge.secureMaterial, key)
}
func (challenge *PresentedChallenge) Wire() ([]byte, error) { return challenge.copy() }

func challengeHMAC(material *secureMaterial, key []byte) ([32]byte, error) {
	if len(key) != 32 {
		return [32]byte{}, authentication.ErrAuthenticationContract
	}
	wire, err := material.copy()
	if err != nil {
		return [32]byte{}, authentication.ErrAuthenticationContract
	}
	defer clear(wire)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("harpy:iam-auth-challenge:v1\x00"))
	_, _ = mac.Write(wire)
	var result [32]byte
	copy(result[:], mac.Sum(nil))
	return result, nil
}
