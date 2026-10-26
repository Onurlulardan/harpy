package licenseauthority

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func ParseTrustBundle(
	ctx context.Context, document []byte, now time.Time,
) (TrustBundle, error) {
	if ctx == nil || ctx.Err() != nil || now.IsZero() || now.Location() != time.UTC ||
		len(document) == 0 || len(document) > MaximumDocumentBytes {
		return TrustBundle{}, ErrInvalidTrustBundle
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
	if err != nil || !bytes.Equal(document, canonical) {
		return TrustBundle{}, ErrInvalidTrustBundle
	}
	var wire trustBundleWire
	if strictDecode(canonical, &wire) != nil || wire.Schema != TrustBundleSchema ||
		wire.Purpose != EnvelopeSchema || len(wire.Roots) == 0 || len(wire.Roots) > 8 ||
		len(wire.Keysets) == 0 || len(wire.Keysets) > 32 {
		return TrustBundle{}, ErrInvalidTrustBundle
	}
	roots, err := parseTrustRoots(wire.Roots)
	if err != nil {
		return TrustBundle{}, err
	}
	bundle := TrustBundle{issuers: make(map[string]trustedIssuer, len(wire.Keysets))}
	for index := range wire.Keysets {
		issuer, parsedErr := verifyKeyset(wire.Keysets[index], roots, now)
		if parsedErr != nil {
			return TrustBundle{}, parsedErr
		}
		if overlapsRootKey(issuer, roots) {
			return TrustBundle{}, ErrInvalidTrustBundle
		}
		if _, duplicate := bundle.issuers[issuer.name]; duplicate {
			return TrustBundle{}, ErrInvalidTrustBundle
		}
		bundle.issuers[issuer.name] = issuer.trustedIssuer
	}
	digest := sha256.Sum256(canonical)
	bundle.digestHex = hex.EncodeToString(digest[:])
	return bundle, nil
}

func parseTrustRoots(values []trustRootWire) (map[string]trustedLicenseKey, error) {
	roots := make(map[string]trustedLicenseKey, len(values))
	previous := ""
	for _, value := range values {
		if previous != "" && value.KeyID <= previous {
			return nil, ErrInvalidTrustBundle
		}
		key, err := parseTrustKey(value.KeyID, value.Algorithm, value.PublicKey,
			value.Status, value.NotBefore, value.SignUntil, value.VerifyUntil)
		if err != nil {
			return nil, err
		}
		roots[value.KeyID], previous = key, value.KeyID
	}
	return roots, nil
}

func overlapsRootKey(issuer parsedTrustedIssuer, roots map[string]trustedLicenseKey) bool {
	for _, signer := range issuer.keys {
		for _, root := range roots {
			if bytes.Equal(signer.publicKey, root.publicKey) {
				return true
			}
		}
	}
	return false
}

var _ KeyResolver = TrustBundle{}

func (bundle TrustBundle) LicenseKey(
	_ context.Context, issuer, keyID string, issuedAt, verifiedAt time.Time,
) (ed25519.PublicKey, error) {
	trusted, found := bundle.issuers[issuer]
	if !found {
		return nil, ErrUntrustedKey
	}
	key, found := trusted.keys[keyID]
	if !found || !keyCanVerify(key, issuedAt, verifiedAt) {
		return nil, ErrUntrustedKey
	}
	return append(ed25519.PublicKey(nil), key.publicKey...), nil
}
