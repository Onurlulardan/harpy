package attestation

import (
	"crypto/ed25519"
	"fmt"
	"regexp"
	"time"
)

var keyIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

type TrustRoot struct {
	keys map[string]trustedKey
}

type trustedKey struct {
	publicKey ed25519.PublicKey
	notBefore time.Time
	notAfter  time.Time
}

func ParseTrustRoot(document []byte) (TrustRoot, error) {
	var decoded trustRootDocument
	if err := decodeStrict(document, &decoded); err != nil {
		return TrustRoot{}, err
	}
	if decoded.ContractVersion != "attestation-trust-root.v1" ||
		len(decoded.Keys) == 0 || len(decoded.Keys) > 32 {
		return TrustRoot{}, fmt.Errorf("attestation trust root contract is invalid")
	}
	root := TrustRoot{keys: make(map[string]trustedKey, len(decoded.Keys))}
	for _, record := range decoded.Keys {
		if _, duplicate := root.keys[record.KeyID]; duplicate {
			return TrustRoot{}, fmt.Errorf("attestation trust root has duplicate key ID")
		}
		key, err := parseTrustedKey(record)
		if err != nil {
			return TrustRoot{}, err
		}
		root.keys[record.KeyID] = key
	}
	return root, nil
}

func parseTrustedKey(record trustedKeyRecord) (trustedKey, error) {
	if !keyIDPattern.MatchString(record.KeyID) || record.Algorithm != algorithmEd25519 {
		return trustedKey{}, fmt.Errorf("attestation trusted key identity is invalid")
	}
	publicKey, err := decodeBase64URLExact(record.PublicKey, ed25519.PublicKeySize)
	if err != nil {
		return trustedKey{}, fmt.Errorf("attestation trusted public key is invalid")
	}
	notBefore, err := parseUTCTimestamp(record.NotBefore)
	if err != nil {
		return trustedKey{}, err
	}
	notAfter, err := parseUTCTimestamp(record.NotAfter)
	if err != nil || !notBefore.Before(notAfter) {
		return trustedKey{}, fmt.Errorf("attestation trusted key validity is invalid")
	}
	return trustedKey{
		publicKey: ed25519.PublicKey(publicKey), notBefore: notBefore, notAfter: notAfter,
	}, nil
}

func parseUTCTimestamp(value string) (time.Time, error) {
	if len(value) == 0 || value[len(value)-1] != 'Z' {
		return time.Time{}, fmt.Errorf("attestation timestamp must be UTC")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("attestation timestamp is invalid")
	}
	return parsed, nil
}
